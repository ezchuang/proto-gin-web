package presenter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/platform/config"
	platformview "proto-gin-web/internal/platform/http/view"
)

func AdminLoginPage(c *gin.Context, cfg config.Config, errMsg string) {
	platformview.RenderHTML(c, http.StatusOK, "admin_login.tmpl", platformview.WithAdminContext(c, gin.H{
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"Error":           errMsg,
	}))
}

func AdminRegisterPage(c *gin.Context, cfg config.Config, errMsg string) {
	platformview.RenderHTML(c, http.StatusOK, "admin_register.tmpl", platformview.WithAdminContext(c, gin.H{
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"Error":           errMsg,
	}))
}

