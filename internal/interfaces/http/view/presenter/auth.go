package presenter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/http/view"
)

func AdminLoginPage(c *gin.Context, cfg platform.Config, errMsg string) {
	view.RenderHTML(c, http.StatusOK, "admin_login.tmpl", gin.H{
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"Error":           errMsg,
	})
}

func AdminRegisterPage(c *gin.Context, cfg platform.Config, errMsg string) {
	view.RenderHTML(c, http.StatusOK, "admin_register.tmpl", gin.H{
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"Error":           errMsg,
	})
}
