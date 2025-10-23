package http

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	bf "github.com/russross/blackfriday/v2"

	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/infrastructure/feed"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/infrastructure/seo"
)

// registerPublicRoutes mounts health checks, SEO endpoints, and SSR pages.
func registerPublicRoutes(r *gin.Engine, cfg platform.Config, postSvc domain.PostService) {
	logger := slog.Default()
	respondInternal := func(c *gin.Context, msg string, err error) {
		logger.Error(msg,
			slog.String("request_id", GetRequestID(c)),
			slog.String("path", c.FullPath()),
			slog.Any("error", err))
		c.String(http.StatusInternalServerError, "internal server error")
	}

	r.GET("/livez", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if _, err := postSvc.ListPublished(ctx, domain.ListPostsOptions{Limit: 1}); err != nil {
			logger.Warn("readiness probe failed",
				slog.String("request_id", GetRequestID(c)),
				slog.Any("error", err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	r.GET("/", func(c *gin.Context) {
		m := seo.Default(cfg.SiteName, cfg.SiteDescription, cfg.BaseURL)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
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
	})

	r.GET("/robots.txt", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusOK, "User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", cfg.BaseURL)
	})

	r.GET("/sitemap.xml", func(c *gin.Context) {
		ctx := c.Request.Context()
		rows, err := postSvc.ListPublished(ctx, domain.ListPostsOptions{Limit: 100})
		if err != nil {
			respondInternal(c, "list published posts for sitemap failed", err)
			return
		}
		paths := []string{"/"}
		for _, p := range rows {
			paths = append(paths, "/posts/"+p.Slug)
		}
		xmlBytes, err := seo.BuildFromPaths(cfg.BaseURL, paths)
		if err != nil {
			respondInternal(c, "build sitemap failed", err)
			return
		}
		c.Header("Content-Type", "application/xml; charset=utf-8")
		c.Writer.Write(xmlBytes)
	})

	r.GET("/rss.xml", func(c *gin.Context) {
		ctx := c.Request.Context()
		rows, err := postSvc.ListPublished(ctx, domain.ListPostsOptions{Limit: 20})
		if err != nil {
			respondInternal(c, "list published posts for rss failed", err)
			return
		}
		base := strings.TrimRight(cfg.BaseURL, "/")
		var items []feed.Item
		for _, p := range rows {
			link := base + "/posts/" + p.Slug
			items = append(items, feed.Item{
				Title:       p.Title,
				Link:        link,
				Description: p.Summary,
				PubDate:     p.PublishedAt,
			})
		}
		xmlBytes, err := feed.BuildRSS(feed.Channel{
			Title:       cfg.SiteName,
			Link:        cfg.BaseURL,
			Description: cfg.SiteDescription,
		}, items)
		if err != nil {
			respondInternal(c, "build rss feed failed", err)
			return
		}
		c.Header("Content-Type", "application/rss+xml; charset=utf-8")
		c.Writer.Write(xmlBytes)
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
		rows, err := postSvc.ListPublished(ctx, domain.ListPostsOptions{
			Category: category,
			Tag:      tag,
			Sort:     sort,
			Limit:    int32(size),
			Offset:   int32(offset),
		})
		if err != nil {
			respondInternal(c, "list posts for index failed", err)
			return
		}
		m := seo.Default(cfg.SiteName, cfg.SiteDescription, cfg.BaseURL).WithPage("Posts", cfg.SiteDescription, cfg.BaseURL+"/posts", "")
		c.HTML(http.StatusOK, "posts.tmpl", gin.H{
			"Title":           "Posts",
			"Env":             cfg.Env,
			"BaseURL":         cfg.BaseURL,
			"SiteName":        cfg.SiteName,
			"SiteDescription": cfg.SiteDescription,
			"Posts":           rows,
			"Page":            page,
			"Size":            size,
			"MetaTags":        template.HTML(m.Tags()),
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

		md := result.Post.ContentMD
		md = strings.ReplaceAll(md, "\r\n", "\n")
		md = strings.ReplaceAll(md, "\\r\\n", "\n")
		md = strings.ReplaceAll(md, "\\n", "\n")
		unsafe := bf.Run([]byte(md))
		safe := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

		m := seo.Default(cfg.SiteName, cfg.SiteDescription, cfg.BaseURL).WithPage(result.Post.Title, result.Post.Summary, cfg.BaseURL+"/posts/"+slug, result.Post.CoverURL)
		// Mark as article for richer previews
		m.Type = "article"
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
			"MetaTags":        template.HTML(m.Tags()),
		})
	})
}
