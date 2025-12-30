package public

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	postdomain "proto-gin-web/internal/blog/post/domain"
	postusecase "proto-gin-web/internal/application/post"
	"proto-gin-web/internal/platform/config"
	"proto-gin-web/internal/platform/seo"
)

func registerSEORoutes(r *gin.Engine, cfg config.Config, postSvc postusecase.PostService) {
	r.GET("/robots.txt", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusOK, "User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", cfg.BaseURL)
	})

	r.GET("/sitemap.xml", func(c *gin.Context) {
		ctx := c.Request.Context()
		rows, err := postSvc.ListPublished(ctx, postdomain.ListPostsOptions{Limit: 100})
		if err != nil {
			c.String(http.StatusInternalServerError, "internal server error")
			return
		}
		paths := []string{"/"}
		for _, p := range rows {
			paths = append(paths, "/posts/"+p.Slug)
		}
		xmlBytes, err := seo.BuildFromPaths(cfg.BaseURL, paths)
		if err != nil {
			c.String(http.StatusInternalServerError, "internal server error")
			return
		}
		c.Header("Content-Type", "application/xml; charset=utf-8")
		c.Writer.Write(xmlBytes)
	})

	r.GET("/rss.xml", func(c *gin.Context) {
		ctx := c.Request.Context()
		rows, err := postSvc.ListPublished(ctx, postdomain.ListPostsOptions{Limit: 20})
		if err != nil {
			c.String(http.StatusInternalServerError, "internal server error")
			return
		}
		base := strings.TrimRight(cfg.BaseURL, "/")
		var (
			items     []seo.RSSItem
			latestPub *time.Time
		)
		for _, p := range rows {
			link := base + "/posts/" + p.Slug
			items = append(items, seo.RSSItem{
				Title:           p.Title,
				Link:            link,
				Description:     p.Summary,
				PubDate:         p.PublishedAt,
				GUID:            link,
				GUIDIsPermaLink: true,
			})
			if p.PublishedAt != nil {
				if latestPub == nil || p.PublishedAt.After(*latestPub) {
					latestPub = p.PublishedAt
				}
			}
		}
		now := time.Now()
		rss := seo.RSS{
			Title:         cfg.SiteName,
			Link:          cfg.BaseURL,
			Description:   cfg.SiteDescription,
			LastBuildDate: &now,
			Items:         items,
		}
		if latestPub != nil {
			rss.PubDate = latestPub
		}
		xmlBytes, err := rss.Build()
		if err != nil {
			c.String(http.StatusInternalServerError, "internal server error")
			return
		}
		c.Header("Content-Type", "application/rss+xml; charset=utf-8")
		c.Writer.Write(xmlBytes)
	})
}

