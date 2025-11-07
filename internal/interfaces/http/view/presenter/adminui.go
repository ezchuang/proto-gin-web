package presenter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/http/view"
)

// AdminPostsPage renders the admin posts list.
func AdminPostsPage(c *gin.Context, cfg platform.Config, posts []domain.Post) {
	view.RenderHTML(c, http.StatusOK, "admin_posts.tmpl", gin.H{
		"Title":           "Admin · Posts · " + cfg.SiteName,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Posts":           posts,
	})
}

// AdminPostFormNew renders the new post form.
func AdminPostFormNew(c *gin.Context, cfg platform.Config) {
	view.RenderHTML(c, http.StatusOK, "admin_post_form.tmpl", gin.H{
		"Title":           "Admin · New Post · " + cfg.SiteName,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"IsNew":           true,
	})
}

// AdminPostFormEdit renders the edit post form.
func AdminPostFormEdit(c *gin.Context, cfg platform.Config, result domain.PostWithRelations) {
	view.RenderHTML(c, http.StatusOK, "admin_post_form.tmpl", gin.H{
		"Title":           "Admin · Edit Post · " + result.Post.Title + " · " + cfg.SiteName,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"IsNew":           false,
		"Post":            result.Post,
		"Categories":      result.Categories,
		"Tags":            result.Tags,
	})
}
