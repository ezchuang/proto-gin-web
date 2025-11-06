package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/http/view"
)

func registerDashboardRoutes(group *gin.RouterGroup, cfg platform.Config) {
	group.GET("", func(c *gin.Context) {
		userName := cfg.AdminUser
		if v, err := c.Cookie("admin_user"); err == nil && v != "" {
			userName = v
		}
		view.RenderHTML(c, http.StatusOK, "admin_dashboard.tmpl", gin.H{
			"SiteName":        cfg.SiteName,
			"SiteDescription": cfg.SiteDescription,
			"Env":             cfg.Env,
			"BaseURL":         cfg.BaseURL,
			"User":            userName,
			"Registered":      c.Query("registered") == "1",
		})
	})
}
