package presenter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	postdomain "proto-gin-web/internal/blog/post/domain"
	"proto-gin-web/internal/infrastructure/platform"
	platformview "proto-gin-web/internal/platform/http/view"
)

// AdminPostsPage renders the admin posts list.
func AdminPostsPage(c *gin.Context, cfg platform.Config, posts []postdomain.Post) {
	platformview.RenderHTML(c, http.StatusOK, "admin_posts.tmpl", platformview.WithAdminContext(c, gin.H{
		"Title":           "Admin · Posts · " + cfg.SiteName,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Posts":           posts,
	}))
}

// AdminPostFormNew renders the new post form.
func AdminPostFormNew(c *gin.Context, cfg platform.Config) {
	platformview.RenderHTML(c, http.StatusOK, "admin_post_form.tmpl", platformview.WithAdminContext(c, gin.H{
		"Title":           "Admin · New Post · " + cfg.SiteName,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"IsNew":           true,
	}))
}

// AdminPostFormEdit renders the edit post form.
func AdminPostFormEdit(c *gin.Context, cfg platform.Config, result postdomain.PostWithRelations) {
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
	}))
}
