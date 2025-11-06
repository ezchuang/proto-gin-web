package http

import (
	"html/template"
	"time"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/infrastructure/platform"
	helper "proto-gin-web/internal/interfaces/http/templates"
)

// NewRouter wires middleware, views, and routes.
func NewRouter(cfg platform.Config, postSvc domain.PostService, adminSvc domain.AdminService, taxonomySvc domain.TaxonomyService) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(SecurityHeaders())
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

	r.HTMLRender = helper.LoadTemplates("internal/interfaces/http/views", "layouts/*.tmpl", "includes/*.tmpl")
	r.Static("/static", "web/static")

	registerPublicRoutes(r, cfg, postSvc)
	registerAPIRoutes(r, postSvc)
	registerAdminRoutes(r, cfg, postSvc, adminSvc, taxonomySvc)

	return r
}
