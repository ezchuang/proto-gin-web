package presenter

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/infrastructure/seo"
	"proto-gin-web/internal/interfaces/http/view"
)

// PublicLanding renders the site landing page.
func PublicLanding(c *gin.Context, cfg platform.Config) {
	m := seo.Default(cfg.SiteName, cfg.SiteDescription, cfg.BaseURL)
	view.RenderHTML(c, http.StatusOK, "index.tmpl", gin.H{
		"Title":           "Index",
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"DocsURL":         "https://gin-gonic.com/en/docs/",
		"PostsURL":        "/posts",
		"APIPostsURL":     "/api/posts?limit=10&offset=0",
		"MetaTags":        template.HTML(m.Tags()),
	})
}

// PublicPosts renders the paginated posts list.
func PublicPosts(c *gin.Context, cfg platform.Config, posts []domain.Post, page, size int64) {
	m := seo.Default(cfg.SiteName, cfg.SiteDescription, cfg.BaseURL).WithPage("Posts", cfg.SiteDescription, cfg.BaseURL+"/posts", "")
	view.RenderHTML(c, http.StatusOK, "posts.tmpl", gin.H{
		"Title":           "Posts",
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Posts":           posts,
		"Page":            page,
		"Size":            size,
		"MetaTags":        template.HTML(m.Tags()),
	})
}

// PublicPostDetail renders a single post detail page.
func PublicPostDetail(c *gin.Context, cfg platform.Config, post domain.PostWithRelations, content template.HTML) {
	m := seo.Default(cfg.SiteName, cfg.SiteDescription, cfg.BaseURL).
		WithPage(post.Post.Title, post.Post.Summary, cfg.BaseURL+"/posts/"+post.Post.Slug, post.Post.CoverURL)
	m.Type = "article"
	view.RenderHTML(c, http.StatusOK, "post.tmpl", gin.H{
		"Title":           post.Post.Title,
		"Summary":         post.Post.Summary,
		"CoverURL":        post.Post.CoverURL,
		"ContentHTML":     content,
		"Categories":      post.Categories,
		"Tags":            post.Tags,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"MetaTags":        template.HTML(m.Tags()),
	})
}
