package responder

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"log/slog"
)

// JSONSuccess wraps payload with ok=true to keep API responses consistent.
func JSONSuccess(c *gin.Context, status int, payload any) {
	if status == 0 {
		status = http.StatusOK
	}
	c.JSON(status, gin.H{
		"ok":   true,
		"data": payload,
	})
}

// JSONError emits a standard error envelope.
func JSONError(c *gin.Context, status int, message string) {
	if status == 0 {
		status = http.StatusInternalServerError
	}
	logJSONError(c, status, message)
	c.JSON(status, gin.H{
		"ok":    false,
		"error": message,
	})
}

func logJSONError(c *gin.Context, status int, message string) {
	logger := slog.Default()
	rid := c.Writer.Header().Get("X-Request-ID")
	if rid == "" {
		rid = c.GetHeader("X-Request-ID")
	}
	args := []any{
		slog.Int("status", status),
		slog.String("error", message),
	}
	if rid != "" {
		args = append(args, slog.String("request_id", rid))
	}
	if req := c.Request; req != nil {
		args = append(args, slog.String("method", req.Method))
		if req.URL != nil {
			args = append(args, slog.String("path", req.URL.Path))
		}
	}
	if status >= http.StatusInternalServerError {
		logger.Error("http error response", args...)
		return
	}
	logger.Warn("http error response", args...)
}
