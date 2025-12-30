package http

import (
	"html/template"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	swaggerdocs "proto-gin-web/docs"
	authsession "proto-gin-web/internal/contexts/admin/auth/session"
	contenthttp "proto-gin-web/internal/contexts/admin/content/adapters/http"
	admincontentusecase "proto-gin-web/internal/contexts/admin/content/usecase"
	adminroutes "proto-gin-web/internal/contexts/admin/ui/adapters/http"
	adminuiusecase "proto-gin-web/internal/contexts/admin/ui/usecase"
	adminusecase "proto-gin-web/internal/contexts/admin/auth/usecase"
	apiroutes "proto-gin-web/internal/contexts/blog/post/adapters/api"
	publicroutes "proto-gin-web/internal/contexts/blog/post/adapters/public"
	postusecase "proto-gin-web/internal/contexts/blog/post/usecase"
	"proto-gin-web/internal/platform/config"
	helper "proto-gin-web/internal/platform/http/templates"
)

// NewRouter wires middleware, templates, and routes.
func NewRouter(cfg config.Config, postSvc postusecase.PostService, adminSvc adminusecase.AdminService, adminContentSvc *admincontentusecase.Service, adminUISvc *adminuiusecase.Service, sessionMgr *authsession.Manager) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	swaggerdocs.SwaggerInfo.Title = cfg.SiteName + " API"
	swaggerdocs.SwaggerInfo.Description = cfg.SiteDescription
	swaggerdocs.SwaggerInfo.BasePath = "/"
	swaggerdocs.SwaggerInfo.Host = hostFromBaseURL(cfg.BaseURL)
	if cfg.Env == "production" {
		swaggerdocs.SwaggerInfo.Schemes = []string{"https", "http"}
	} else {
		swaggerdocs.SwaggerInfo.Schemes = []string{"http", "https"}
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

	r.HTMLRender = helper.LoadTemplates("internal/platform/http/templates", "layouts/*.tmpl", "includes/*.tmpl")
	r.Static("/static", "web/static")

	publicroutes.RegisterRoutes(r, cfg, postSvc)
	apiroutes.RegisterRoutes(r, postSvc)
	loginLimiter := NewIPRateLimiter(5, time.Minute)
	registerLimiter := NewIPRateLimiter(3, time.Minute)
	sessionGuard := AdminAuth(cfg, sessionMgr, adminSvc)
	adminroutes.RegisterRoutes(r, cfg, adminSvc, sessionMgr, loginLimiter, registerLimiter, sessionGuard)
	apiGroup := r.Group("/admin")
	apiGroup.Use(sessionGuard)
	contenthttp.RegisterRoutes(apiGroup, adminContentSvc)
	adminroutes.RegisterUIRoutes(r, cfg, adminSvc, adminUISvc, sessionMgr, sessionGuard)
	if cfg.Env != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return r
}

func hostFromBaseURL(base string) string {
	if base == "" {
		return "localhost:8080"
	}
	u, err := url.Parse(base)
	if err == nil && u.Host != "" {
		return u.Host
	}
	trimmed := strings.TrimPrefix(base, "https://")
	trimmed = strings.TrimPrefix(trimmed, "http://")
	if trimmed == "" {
		return "localhost:8080"
	}
	return trimmed
}





