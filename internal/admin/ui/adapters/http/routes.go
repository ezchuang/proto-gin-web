package adminuihttp

import (
	"github.com/gin-gonic/gin"

	authhttp "proto-gin-web/internal/admin/auth/adapters/http"
	authdomain "proto-gin-web/internal/admin/auth/domain"
	contenthttp "proto-gin-web/internal/admin/content/adapters/http"
	admincontentusecase "proto-gin-web/internal/application/admincontent"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/auth"
)

// RegisterRoutes wires admin authentication and CRUD handlers.
func RegisterRoutes(r *gin.Engine, cfg platform.Config, adminSvc authdomain.AdminService, contentSvc *admincontentusecase.Service, loginLimiter, registerLimiter gin.HandlerFunc) {
	authhttp.RegisterRoutes(r, cfg, adminSvc, loginLimiter, registerLimiter)
	adminGroup := r.Group("/admin", auth.AdminRequired())
	registerDashboardRoutes(adminGroup, cfg)
	registerProfileRoutes(adminGroup, cfg, adminSvc)
	contenthttp.RegisterRoutes(adminGroup, contentSvc)
}
