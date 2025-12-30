package public

import (
	"github.com/gin-gonic/gin"

	postdomain "proto-gin-web/internal/blog/post/domain"
	"proto-gin-web/internal/platform/config"
)

// RegisterRoutes wires all public-facing routes.
func RegisterRoutes(r *gin.Engine, cfg config.Config, postSvc postdomain.PostService) {
	registerHealthRoutes(r, postSvc)
	registerSEORoutes(r, cfg, postSvc)
	registerContentRoutes(r, cfg, postSvc)
}

