package presenter

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"

	postdomain "proto-gin-web/internal/blog/post/domain"
	"proto-gin-web/internal/infrastructure/platform"
	platformview "proto-gin-web/internal/platform/http/view"
	"proto-gin-web/internal/platform/seo"
)

// PublicLanding renders the site landing page.
func PublicLanding(c *gin.Context, cfg platform.Config) {
	m := seo.Default(cfg.SiteName, cfg.SiteDescription, cfg.BaseURL)
	data := gin.H{
		"Title":           "Index",
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"DocsURL":         "https://gin-gonic.com/en/docs/",
		"PostsURL":        "/posts",
		"APIPostsURL":     "/api/posts?limit=10&offset=0",
		"AdminUIURL":      "/admin/ui/posts",
		"AdminNewPostURL": "/admin/ui/posts/new",
		"LivezURL":        "/livez",
		"ReadyzURL":       "/readyz",
		"SwaggerURL":      "",
		"MetaTags":        template.HTML(m.Tags()),
	}
	if cfg.Env != "production" {
		data["SwaggerURL"] = "/swagger/index.html"
	}
	platformview.RenderHTML(c, http.StatusOK, "index.tmpl", platformview.WithAdminContext(c, data))
}

// PublicPosts renders the paginated posts list.
func PublicPosts(c *gin.Context, cfg platform.Config, posts []postdomain.Post, page, size int64) {
	m := seo.Default(cfg.SiteName, cfg.SiteDescription, cfg.BaseURL).WithPage("Posts", cfg.SiteDescription, cfg.BaseURL+"/posts", "")
	platformview.RenderHTML(c, http.StatusOK, "posts.tmpl", platformview.WithAdminContext(c, gin.H{
		"Title":           "Posts",
		"Env":             cfg.Env,
		"BaseURL":         cfg.BaseURL,
		"SiteName":        cfg.SiteName,
		"SiteDescription": cfg.SiteDescription,
		"Posts":           posts,
		"Page":            page,
		"Size":            size,
		"MetaTags":        template.HTML(m.Tags()),
	}))
}

// PublicPostDetail renders a single post detail page.
func PublicPostDetail(c *gin.Context, cfg platform.Config, post postdomain.PostWithRelations, content template.HTML) {
	m := seo.Default(cfg.SiteName, cfg.SiteDescription, cfg.BaseURL).
		WithPage(post.Post.Title, post.Post.Summary, cfg.BaseURL+"/posts/"+post.Post.Slug, post.Post.CoverURL)
	m.Type = "article"
	platformview.RenderHTML(c, http.StatusOK, "post.tmpl", platformview.WithAdminContext(c, gin.H{
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
	}))
}
