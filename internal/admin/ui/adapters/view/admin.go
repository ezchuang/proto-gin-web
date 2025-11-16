package presenter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	authdomain "proto-gin-web/internal/admin/auth/domain"
	"proto-gin-web/internal/infrastructure/platform"
	platformview "proto-gin-web/internal/platform/http/view"
)

func AdminDashboard(c *gin.Context, cfg platform.Config, user string, registered bool) {
	platformview.RenderHTML(c, http.StatusOK, "admin_dashboard.tmpl", platformview.WithAdminContext(c, gin.H{
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"User":            user,
		"Registered":      registered,
	}))
}

func AdminProfilePage(c *gin.Context, cfg platform.Config, profile authdomain.Admin, updated bool, errMsg string) {
	platformview.RenderHTML(c, http.StatusOK, "admin_profile.tmpl", platformview.WithAdminContext(c, gin.H{
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
	platformview.RenderHTML(c, status, "admin_profile.tmpl", platformview.WithAdminContext(c, gin.H{
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
