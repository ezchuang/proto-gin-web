package public

import (
	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/infrastructure/platform"
)

// RegisterRoutes wires all public-facing routes.
func RegisterRoutes(r *gin.Engine, cfg platform.Config, postSvc domain.PostService) {
	registerHealthRoutes(r, postSvc)
	registerSEORoutes(r, cfg, postSvc)
	registerContentRoutes(r, cfg, postSvc)
}
