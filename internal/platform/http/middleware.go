package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	authdomain "proto-gin-web/internal/admin/auth/domain"
)

type requestIDKey struct{}

const requestIDContextKey = "request_id"

// RequestID assigns/propagates a request identifier for tracing across logs and responses.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = generateRequestID()
		}
		ctx := context.WithValue(c.Request.Context(), requestIDKey{}, rid)
		c.Request = c.Request.WithContext(ctx)
		c.Set(requestIDContextKey, rid)
		c.Writer.Header().Set("X-Request-ID", rid)
		c.Next()
	}
}

// RequestLogger emits structured logs with core metadata about the request.
func RequestLogger() gin.HandlerFunc {
	logger := slog.Default()
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		rid := GetRequestID(c)
		logger.Info("http request",
			slog.String("request_id", rid),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.String("client_ip", c.ClientIP()),
			slog.Duration("latency", latency),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.Any("errors", c.Errors),
		)
	}
}

// RecoveryWithRequestID recovers from panics, logging the request and surfacing request-id in the response header.
func RecoveryWithRequestID() gin.HandlerFunc {
	logger := slog.Default()
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				rid := GetRequestID(c)
				if rid != "" {
					c.Writer.Header().Set("X-Request-ID", rid)
				}
				logger.Error("panic recovered",
					slog.Any("error", rec),
					slog.String("request_id", rid),
					slog.String("path", c.Request.URL.Path),
					slog.String("method", c.Request.Method),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":      "internal server error",
					"request_id": rid,
				})
			}
		}()
		c.Next()
	}
}

// GetRequestID fetches the request identifier stored on the current context.
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(requestIDContextKey); ok {
		if rid, ok := v.(string); ok {
			return rid
		}
	}
	if v := c.Request.Context().Value(requestIDKey{}); v != nil {
		if rid, ok := v.(string); ok {
			return rid
		}
	}
	return ""
}

func generateRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// SecurityHeaders adds common security-focused headers to every response.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		// CSP 暫時關閉以便 Swagger 等內嵌腳本正常運作。
		// if _, ok := h["Content-Security-Policy"]; !ok {
		// 	h.Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; base-uri 'self'")
		// }
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "same-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Next()
	}
}

// NewIPRateLimiter limits requests per client IP within the provided window.
func NewIPRateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	type entry struct {
		count  int
		expiry time.Time
	}
	var (
		mu    sync.Mutex
		store = make(map[string]entry)
	)
	logger := slog.Default()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		e := store[ip]
		if now.After(e.expiry) {
			e = entry{count: 0, expiry: now.Add(window)}
		}
		e.count++
		store[ip] = e
		mu.Unlock()

		if e.count > maxRequests {
			logger.Warn("rate limit exceeded",
				slog.String("request_id", GetRequestID(c)),
				slog.String("ip", ip),
				slog.String("path", c.Request.URL.Path),
				slog.Int("max_requests", maxRequests),
				slog.Duration("window", window))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many requests, try again later",
				"retry_after": int(e.expiry.Sub(now).Seconds()),
				"request_id":  GetRequestID(c),
			})
			return
		}

		c.Next()
	}
}

// NOTE: For high-traffic endpoints, consider replacing this mutex-backed map with a sharded map,
// atomic counters, or a dedicated rate-limiting backend (e.g., Redis) to avoid contention.

type adminProfileKey struct{}

// AdminAuth validates the admin session cookie and surfaces the admin profile in context.
func AdminAuth(adminSvc authdomain.AdminService) gin.HandlerFunc {
	logger := slog.Default()
	return func(c *gin.Context) {
		email, err := c.Cookie("admin_email")
		if err != nil || strings.TrimSpace(email) == "" {
			abortUnauthorized(c)
			return
		}
		profile, err := adminSvc.GetProfile(c.Request.Context(), strings.TrimSpace(email))
		if err != nil {
			if errors.Is(err, authdomain.ErrAdminNotFound) {
				logger.Warn("admin auth profile missing",
					slog.String("email", email),
					slog.String("request_id", GetRequestID(c)))
			} else {
				logger.Error("admin auth lookup failed",
					slog.Any("err", err),
					slog.String("request_id", GetRequestID(c)))
			}
			abortUnauthorized(c)
			return
		}
		ctx := context.WithValue(c.Request.Context(), adminProfileKey{}, profile)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// AdminProfileFromContext extracts the authenticated admin profile when AdminAuth ran.
func AdminProfileFromContext(c *gin.Context) (authdomain.Admin, bool) {
	if v := c.Request.Context().Value(adminProfileKey{}); v != nil {
		if profile, ok := v.(authdomain.Admin); ok {
			return profile, true
		}
	}
	return authdomain.Admin{}, false
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error":      "unauthorized",
		"request_id": GetRequestID(c),
	})
}
