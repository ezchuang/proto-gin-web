package public

import (
	"github.com/gin-gonic/gin"

	postusecase "proto-gin-web/internal/application/post"
	"proto-gin-web/internal/platform/config"
)

// RegisterRoutes wires all public-facing routes.
func RegisterRoutes(r *gin.Engine, cfg config.Config, postSvc postusecase.PostService) {
	registerHealthRoutes(r, postSvc)
	registerSEORoutes(r, cfg, postSvc)
	registerContentRoutes(r, cfg, postSvc)
}

