package http

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/auth"
)

// registerAdminRoutes wires admin authentication and CRUD handlers.
func registerAdminRoutes(r *gin.Engine, cfg platform.Config, postSvc domain.PostService, adminSvc domain.AdminService, taxonomySvc domain.TaxonomyService) {
	loginLimiter := NewIPRateLimiter(5, time.Minute)
	registerLimiter := NewIPRateLimiter(3, time.Minute)

	registerAdminAuthRoutes(r, cfg, adminSvc, loginLimiter, registerLimiter)

	adminGroup := r.Group("/admin", auth.AdminRequired())
	registerAdminDashboardRoutes(adminGroup, cfg)
	registerAdminProfileRoutes(adminGroup, cfg, adminSvc)
	registerAdminContentRoutes(adminGroup, postSvc, taxonomySvc)
}

func wantsHTMLResponse(c *gin.Context) bool {
	ct := c.GetHeader("Content-Type")
	if strings.Contains(ct, "application/x-www-form-urlencoded") || strings.Contains(ct, "multipart/form-data") {
		return true
	}
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "text/html")
}
