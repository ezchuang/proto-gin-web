package http

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	authdomain "proto-gin-web/internal/admin/auth/domain"
	authsession "proto-gin-web/internal/admin/auth/session"
	"proto-gin-web/internal/platform/config"
)

type requestIDKey struct{}

const requestIDContextKey = "request_id"
const cspNonceContextKey = "csp_nonce"

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

func generateCSPNonce() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return base64.RawStdEncoding.EncodeToString(buf)
}

func buildCSP(nonce string) string {
	directives := []string{
		"default-src 'self'",
		"base-uri 'self'",
		"frame-ancestors 'none'",
		"object-src 'none'",
		"script-src 'self' 'nonce-" + nonce + "'",
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' https: data:",
		"form-action 'self'",
	}
	return strings.Join(directives, "; ")
}

// SecurityHeaders adds common security-focused headers to every response.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		nonce := generateCSPNonce()
		if nonce != "" {
			c.Set(cspNonceContextKey, nonce)
		}
		if nonce != "" && !strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			if _, ok := h["Content-Security-Policy"]; !ok {
				h.Set("Content-Security-Policy", buildCSP(nonce))
			}
		}
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

// TODO: For high-traffic endpoints, consider replacing this mutex-backed map with a sharded map,
// atomic counters, or a dedicated rate-limiting backend (e.g., Redis) to avoid contention.

type adminProfileKey struct{}

const (
	defaultSessionCookieAge  = 30 * 60
	defaultRememberCookieAge = 30 * 24 * 60 * 60
)

// AdminAuth validates the admin session cookie and surfaces the admin profile in context.
func AdminAuth(cfg config.Config, sessionMgr *authsession.Manager, adminSvc authdomain.AdminService) gin.HandlerFunc {
	logger := slog.Default()
	return func(c *gin.Context) {
		profile, err := resolveAdminSession(c, cfg, sessionMgr, adminSvc)
		if err != nil {
			if errors.Is(err, authdomain.ErrAdminSessionNotFound) {
				logger.Warn("admin auth session missing",
					slog.String("request_id", GetRequestID(c)))
			} else {
				logger.Error("admin auth resolution failed",
					slog.Any("err", err),
					slog.String("request_id", GetRequestID(c)))
			}
			abortUnauthorized(c)
			return
		}
		ctx := context.WithValue(c.Request.Context(), adminProfileKey{}, profile)
		c.Request = c.Request.WithContext(ctx)
		c.Set("admin_profile", profile)
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

func resolveAdminSession(c *gin.Context, cfg config.Config, sessionMgr *authsession.Manager, adminSvc authdomain.AdminService) (authdomain.Admin, error) {
	ctx := c.Request.Context()
	if sessionID, err := c.Cookie(cfg.SessionCookieName); err == nil && strings.TrimSpace(sessionID) != "" {
		session, sessErr := sessionMgr.ValidateSession(ctx, strings.TrimSpace(sessionID))
		if sessErr == nil {
			refreshSessionCookies(c, cfg, session.ID, session.Profile)
			return session.Profile, nil
		}
		if errors.Is(sessErr, authdomain.ErrAdminSessionExpired) || errors.Is(sessErr, authdomain.ErrAdminSessionNotFound) {
			wipeSessionCookie(c, cfg)
		} else if sessErr != nil {
			return authdomain.Admin{}, sessErr
		}
	}
	selector, validator := readRememberCookie(c, cfg)
	if selector == "" || validator == "" {
		return authdomain.Admin{}, authdomain.ErrAdminSessionNotFound
	}
	token, secret, err := sessionMgr.ValidateRememberToken(ctx, selector, validator, normalizeDeviceInfo(c.Request.UserAgent()))
	if err != nil {
		wipeRememberCookie(c, cfg)
		if errors.Is(err, authdomain.ErrAdminRememberTokenInvalid) || errors.Is(err, authdomain.ErrAdminRememberTokenNotFound) {
			return authdomain.Admin{}, authdomain.ErrAdminSessionNotFound
		}
		return authdomain.Admin{}, err
	}
	profile, err := adminSvc.GetProfileByID(ctx, token.UserID)
	if err != nil {
		_ = sessionMgr.DeleteRememberToken(ctx, token.Selector)
		wipeRememberCookie(c, cfg)
		return authdomain.Admin{}, err
	}
	newSession, err := sessionMgr.IssueSession(ctx, profile)
	if err != nil {
		return authdomain.Admin{}, err
	}
	refreshSessionCookies(c, cfg, newSession.ID, profile)
	writeRememberCookie(c, cfg, secret)
	return profile, nil
}

func refreshSessionCookies(c *gin.Context, cfg config.Config, sessionID string, profile authdomain.Admin) {
	secure := cfg.Env == "production"
	c.SetSameSite(http.SameSiteStrictMode)
	maxAge := defaultSessionCookieAge
	c.SetCookie(cfg.SessionCookieName, sessionID, maxAge, "/", "", secure, true)
	// Lightweight display info cookies help SSR templates show current admin context without extra lookups.
	c.SetCookie("admin_user", profile.DisplayName, maxAge, "/", "", secure, true)
	c.SetCookie("admin_email", profile.Email, maxAge, "/", "", secure, true)
}

func wipeSessionCookie(c *gin.Context, cfg config.Config) {
	secure := cfg.Env == "production"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(cfg.SessionCookieName, "", -1, "/", "", secure, true)
}

func readRememberCookie(c *gin.Context, cfg config.Config) (string, string) {
	value, err := c.Cookie(cfg.RememberCookieName)
	if err != nil || value == "" {
		return "", ""
	}
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func writeRememberCookie(c *gin.Context, cfg config.Config, secret authsession.RememberTokenSecret) {
	secure := cfg.Env == "production"
	c.SetSameSite(http.SameSiteStrictMode)
	maxAge := defaultRememberCookieAge
	c.SetCookie(cfg.RememberCookieName, secret.Selector+":"+secret.Validator, maxAge, "/", "", secure, true)
}

func wipeRememberCookie(c *gin.Context, cfg config.Config) {
	secure := cfg.Env == "production"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(cfg.RememberCookieName, "", -1, "/", "", secure, true)
}

func normalizeDeviceInfo(userAgent string) string {
	ua := strings.TrimSpace(userAgent)
	if len(ua) > 255 {
		return ua[:255]
	}
	return ua
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error":      "unauthorized",
		"request_id": GetRequestID(c),
	})
}

