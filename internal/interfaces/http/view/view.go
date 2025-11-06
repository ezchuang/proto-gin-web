package view

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// RenderHTML injects shared template context and renders a view.
func RenderHTML(c *gin.Context, status int, name string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	if adminUser, err := c.Cookie("admin_user"); err == nil && adminUser != "" {
		data["AdminUser"] = adminUser
	}
	if adminEmail, err := c.Cookie("admin_email"); err == nil && adminEmail != "" {
		data["AdminEmail"] = adminEmail
	}
	c.HTML(status, name, data)
}

// WantsHTMLResponse reports whether the caller likely expects HTML output.
func WantsHTMLResponse(c *gin.Context) bool {
	ct := c.GetHeader("Content-Type")
	if strings.Contains(ct, "application/x-www-form-urlencoded") || strings.Contains(ct, "multipart/form-data") {
		return true
	}
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "text/html")
}
