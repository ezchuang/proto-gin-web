package responder

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	c.JSON(status, gin.H{
		"ok":    false,
		"error": message,
	})
}
