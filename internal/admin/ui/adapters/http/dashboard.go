package adminuihttp

import (
	"github.com/gin-gonic/gin"

	adminview "proto-gin-web/internal/admin/ui/adapters/view"
	"proto-gin-web/internal/infrastructure/platform"
)

func registerDashboardRoutes(group *gin.RouterGroup, cfg platform.Config) {
	group.GET("", func(c *gin.Context) {
		userName := "Admin"
		if profile, ok := adminProfileFromContext(c); ok && profile.DisplayName != "" {
			userName = profile.DisplayName
		}
		adminview.AdminDashboard(c, cfg, userName, c.Query("registered") == "1")
	})
}
