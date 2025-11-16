package view

import (
	"strings"

	"github.com/gin-gonic/gin"

	authdomain "proto-gin-web/internal/admin/auth/domain"
)

// RenderHTML renders a view with the provided data.
func RenderHTML(c *gin.Context, status int, name string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	c.HTML(status, name, data)
}

// WithAdminContext attaches admin session info when available.
func WithAdminContext(c *gin.Context, data gin.H) gin.H {
	if data == nil {
		data = gin.H{}
	}
	if profile, ok := c.Get("admin_profile"); ok {
		if admin, ok := profile.(authdomain.Admin); ok {
			if admin.DisplayName != "" {
				data["AdminUser"] = admin.DisplayName
			}
			if admin.Email != "" {
				data["AdminEmail"] = admin.Email
			}
		}
	}
	if adminUser, err := c.Cookie("admin_user"); err == nil && adminUser != "" {
		data["AdminUser"] = adminUser
	}
	if adminEmail, err := c.Cookie("admin_email"); err == nil && adminEmail != "" {
		data["AdminEmail"] = adminEmail
	}
	return data
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
