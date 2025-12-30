package presenter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	postdomain "proto-gin-web/internal/blog/post/domain"
	"proto-gin-web/internal/platform/config"
	platformview "proto-gin-web/internal/platform/http/view"
)

// AdminPostsPage renders the admin posts list.
func AdminPostsPage(c *gin.Context, cfg config.Config, posts []postdomain.Post) {
	platformview.RenderHTML(c, http.StatusOK, "admin_posts.tmpl", platformview.WithAdminContext(c, gin.H{
		"Title":           "Admin · Posts · " + cfg.SiteName,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Posts":           posts,
		"Error":           c.Query("error"),
		"Success":         c.Query("success"),
	}))
}

// AdminPostFormNew renders the new post form.
func AdminPostFormNew(c *gin.Context, cfg config.Config) {
	platformview.RenderHTML(c, http.StatusOK, "admin_post_form.tmpl", platformview.WithAdminContext(c, gin.H{
		"Title":           "Admin · New Post · " + cfg.SiteName,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"IsNew":           true,
		"Error":           c.Query("error"),
		"Success":         c.Query("success"),
	}))
}

// AdminPostFormEdit renders the edit post form.
func AdminPostFormEdit(c *gin.Context, cfg config.Config, result postdomain.PostWithRelations) {
	platformview.RenderHTML(c, http.StatusOK, "admin_post_form.tmpl", platformview.WithAdminContext(c, gin.H{
		"Title":           "Admin · Edit Post · " + result.Post.Title + " · " + cfg.SiteName,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"IsNew":           false,
		"Post":            result.Post,
		"Categories":      result.Categories,
		"Tags":            result.Tags,
		"Error":           c.Query("error"),
		"Success":         c.Query("success"),
	}))
}

