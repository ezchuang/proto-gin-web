package auth

import "github.com/gin-gonic/gin"

// AdminRequired is a minimal middleware that checks an 'admin' cookie.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, err := c.Cookie("admin")
		if err != nil || v != "1" {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
