package admin

import (
	"github.com/gin-gonic/gin"

	admincontentusecase "proto-gin-web/internal/application/admincontent"
	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/auth"
)

// RegisterRoutes wires admin authentication and CRUD handlers.
func RegisterRoutes(r *gin.Engine, cfg platform.Config, adminSvc domain.AdminService, contentSvc *admincontentusecase.Service, loginLimiter, registerLimiter gin.HandlerFunc) {
	registerAuthRoutes(r, cfg, adminSvc, loginLimiter, registerLimiter)
	adminGroup := r.Group("/admin", auth.AdminRequired())
	registerDashboardRoutes(adminGroup, cfg)
	registerProfileRoutes(adminGroup, cfg, adminSvc)
	registerContentRoutes(adminGroup, contentSvc)
}
