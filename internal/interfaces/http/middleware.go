package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
