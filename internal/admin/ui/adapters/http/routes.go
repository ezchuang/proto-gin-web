package adminuihttp

import (
	"github.com/gin-gonic/gin"

	authhttp "proto-gin-web/internal/admin/auth/adapters/http"
	authdomain "proto-gin-web/internal/admin/auth/domain"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/auth"
)

// RegisterRoutes wires admin authentication and CRUD handlers.
func RegisterRoutes(r *gin.Engine, cfg platform.Config, adminSvc authdomain.AdminService, loginLimiter, registerLimiter gin.HandlerFunc) {
	authhttp.RegisterRoutes(r, cfg, adminSvc, loginLimiter, registerLimiter)
	uiGroup := r.Group("/admin", auth.AdminRequired())
	registerDashboardRoutes(uiGroup, cfg)
	registerProfileRoutes(uiGroup, cfg, adminSvc)
}
