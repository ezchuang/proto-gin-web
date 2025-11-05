package http

import "github.com/gin-gonic/gin"

func renderHTML(c *gin.Context, status int, name string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	if adminUser, err := c.Cookie("admin_user"); err == nil && adminUser != "" {
		data["AdminUser"] = adminUser
	}
	c.HTML(status, name, data)
}
