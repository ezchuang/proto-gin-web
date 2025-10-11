package http

import (
	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/core"
	"proto-gin-web/internal/platform"
	appdb "proto-gin-web/internal/repo/pg"
)

// NewRouter wires middleware, views, and routes.
func NewRouter(cfg platform.Config, postSvc core.PostService, queries *appdb.Queries) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(RequestID())
	r.Use(RecoveryWithRequestID())
	r.Use(RequestLogger())

	// Templates & static assets
	r.LoadHTMLGlob("internal/http/views/**/*")
	r.Static("/static", "web/static")

	registerPublicRoutes(r, cfg, postSvc)
	registerAPIRoutes(r, postSvc, queries)
	registerAdminRoutes(r, cfg, postSvc, queries)

	return r
}
