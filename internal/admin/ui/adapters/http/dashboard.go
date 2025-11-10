package adminuihttp

import (
	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/http/view/presenter"
)

func registerDashboardRoutes(group *gin.RouterGroup, cfg platform.Config) {
	group.GET("", func(c *gin.Context) {
		userName := cfg.AdminUser
		if v, err := c.Cookie("admin_user"); err == nil && v != "" {
			userName = v
		}
		presenter.AdminDashboard(c, cfg, userName, c.Query("registered") == "1")
	})
}
