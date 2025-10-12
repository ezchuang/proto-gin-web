package http

import (
	"context"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	bf "github.com/russross/blackfriday/v2"

	"proto-gin-web/internal/core"
	"proto-gin-web/internal/platform"
)

// registerPublicRoutes mounts health checks, SEO endpoints, and SSR pages.
func registerPublicRoutes(r *gin.Engine, cfg platform.Config, postSvc core.PostService) {
	r.GET("/livez", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if _, err := postSvc.ListPublished(ctx, core.ListPostsOptions{Limit: 1}); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Title":           "Proto ??Gin + SSR",
			"Env":             cfg.Env,
			"BaseURL":         cfg.BaseURL,
			"SiteName":        cfg.SiteName,
			"SiteDescription": cfg.SiteDescription,
			"DocsURL":         "https://gin-gonic.com/docs/",
		})
	})

	r.GET("/robots.txt", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusOK, "User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", cfg.BaseURL)
	})

	r.GET("/sitemap.xml", func(c *gin.Context) {
		ctx := c.Request.Context()
		rows, err := postSvc.ListPublished(ctx, core.ListPostsOptions{Limit: 100})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.Header("Content-Type", "application/xml; charset=utf-8")
		c.Writer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
		c.Writer.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
		c.Writer.WriteString("  <url><loc>" + cfg.BaseURL + "/</loc><changefreq>daily</changefreq><priority>1.0</priority></url>\n")
		for _, p := range rows {
			c.Writer.WriteString("  <url><loc>" + cfg.BaseURL + "/posts/" + p.Slug + "</loc></url>\n")
		}
		c.Writer.WriteString("</urlset>")
	})

	r.GET("/rss.xml", func(c *gin.Context) {
		ctx := c.Request.Context()
		rows, err := postSvc.ListPublished(ctx, core.ListPostsOptions{Limit: 20})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.Header("Content-Type", "application/rss+xml; charset=utf-8")
		c.Writer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
		c.Writer.WriteString("<rss version=\"2.0\">\n  <channel>\n")
		c.Writer.WriteString("    <title>" + cfg.SiteName + "</title>\n")
		c.Writer.WriteString("    <link>" + cfg.BaseURL + "</link>\n")
		c.Writer.WriteString("    <description>" + cfg.SiteDescription + "</description>\n")
		for _, p := range rows {
			c.Writer.WriteString("    <item>\n")
			c.Writer.WriteString("      <title>" + p.Title + "</title>\n")
			c.Writer.WriteString("      <link>" + cfg.BaseURL + "/posts/" + p.Slug + "</link>\n")
			c.Writer.WriteString("      <description>" + p.Summary + "</description>\n")
			c.Writer.WriteString("    </item>\n")
		}
		c.Writer.WriteString("  </channel>\n</rss>")
	})

	r.GET("/posts", func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		sizeStr := c.DefaultQuery("size", "10")
		page, _ := strconv.ParseInt(pageStr, 10, 32)
		size, _ := strconv.ParseInt(sizeStr, 10, 32)
		if page < 1 {
			page = 1
		}
		if size < 1 {
			size = 10
		}
		offset := (page - 1) * size

		category := c.Query("category")
		tag := c.Query("tag")
		sort := c.DefaultQuery("sort", "created_at_desc")
		ctx := c.Request.Context()
		rows, err := postSvc.ListPublished(ctx, core.ListPostsOptions{
			Category: category,
			Tag:      tag,
			Sort:     sort,
			Limit:    int32(size),
			Offset:   int32(offset),
		})
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Title":           "Posts ??" + cfg.SiteName,
			"Env":             cfg.Env,
			"BaseURL":         cfg.BaseURL,
			"SiteName":        cfg.SiteName,
			"SiteDescription": cfg.SiteDescription,
			"Posts":           rows,
			"Page":            page,
			"Size":            size,
		})
	})

	r.GET("/posts/:slug", func(c *gin.Context) {
		slug := c.Param("slug")

		ctx := c.Request.Context()
		result, err := postSvc.GetBySlug(ctx, slug)
		if err != nil {
			c.String(http.StatusNotFound, "post not found")
			return
		}

		unsafe := bf.Run([]byte(result.Post.ContentMD))
		safe := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

		c.HTML(http.StatusOK, "post.tmpl", gin.H{
			"Title":           result.Post.Title,
			"Summary":         result.Post.Summary,
			"CoverURL":        result.Post.CoverURL,
			"ContentHTML":     template.HTML(string(safe)),
			"Categories":      result.Categories,
			"Tags":            result.Tags,
			"Env":             cfg.Env,
			"BaseURL":         cfg.BaseURL,
			"SiteName":        cfg.SiteName,
			"SiteDescription": cfg.SiteDescription,
		})
	})
}
