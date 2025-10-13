package http

import (
	"html/template"
	"time"

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

	// Gin template funcs
	r.SetFuncMap(template.FuncMap{
		"timefmt": func(t time.Time, layout ...string) string {
			if len(layout) > 0 && layout[0] != "" {
				return t.Format(layout[0])
			}
			// default ISO 8601-like format
			return t.UTC().Format("2006-01-02T15:04:05Z")
		},
	})

	r.LoadHTMLGlob("internal/http/views/*.tmpl")
	r.Static("/static", "web/static")

	registerPublicRoutes(r, cfg, postSvc)
	registerAPIRoutes(r, postSvc, queries)
	registerAdminRoutes(r, cfg, postSvc, queries)

	return r
}
