package adminuihttp

import (
	"github.com/gin-gonic/gin"

	authhttp "proto-gin-web/internal/admin/auth/adapters/http"
	authdomain "proto-gin-web/internal/admin/auth/domain"
	authsession "proto-gin-web/internal/admin/auth/session"
	"proto-gin-web/internal/infrastructure/platform"
)

// RegisterRoutes wires admin authentication and CRUD handlers.
func RegisterRoutes(r *gin.Engine, cfg platform.Config, adminSvc authdomain.AdminService, sessionMgr *authsession.Manager, loginLimiter, registerLimiter, sessionGuard gin.HandlerFunc) {
	authhttp.RegisterRoutes(r, cfg, adminSvc, sessionMgr, loginLimiter, registerLimiter)
	uiGroup := r.Group("/admin", sessionGuard)
	registerDashboardRoutes(uiGroup, cfg)
	registerProfileRoutes(uiGroup, cfg, adminSvc, sessionMgr)
}
