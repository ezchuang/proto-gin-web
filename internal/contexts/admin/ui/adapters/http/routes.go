package adminuihttp

import (
	"github.com/gin-gonic/gin"

	authhttp "proto-gin-web/internal/contexts/admin/auth/adapters/http"
	authsession "proto-gin-web/internal/contexts/admin/auth/session"
	adminusecase "proto-gin-web/internal/contexts/admin/auth/usecase"
	"proto-gin-web/internal/platform/config"
)

// RegisterRoutes wires admin authentication and CRUD handlers.
func RegisterRoutes(r *gin.Engine, cfg config.Config, adminSvc adminusecase.AdminService, sessionMgr *authsession.Manager, loginLimiter, registerLimiter, sessionGuard gin.HandlerFunc) {
	authhttp.RegisterRoutes(r, cfg, adminSvc, sessionMgr, loginLimiter, registerLimiter)
	uiGroup := r.Group("/admin", sessionGuard)
	registerDashboardRoutes(uiGroup, cfg)
	registerProfileRoutes(uiGroup, cfg, adminSvc, sessionMgr)
}




