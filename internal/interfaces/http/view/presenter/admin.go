package presenter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/http/view"
)

func AdminDashboard(c *gin.Context, cfg platform.Config, user string, registered bool) {
	view.RenderHTML(c, http.StatusOK, "admin_dashboard.tmpl", view.WithAdminContext(c, gin.H{
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"User":            user,
		"Registered":      registered,
	}))
}

func AdminProfilePage(c *gin.Context, cfg platform.Config, profile domain.Admin, updated bool, errMsg string) {
	view.RenderHTML(c, http.StatusOK, "admin_profile.tmpl", view.WithAdminContext(c, gin.H{
		"Title":           "Account Settings",
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"Profile":         profile,
		"Updated":         updated,
		"Error":           errMsg,
	}))
}

func AdminProfileError(c *gin.Context, cfg platform.Config, email, displayName, errMsg string, status int) {
	view.RenderHTML(c, status, "admin_profile.tmpl", view.WithAdminContext(c, gin.H{
		"Title":           "Account Settings",
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"Profile": gin.H{
			"Email":       email,
			"DisplayName": displayName,
		},
		"Error": errMsg,
	}))
}
