package http

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	bf "github.com/russross/blackfriday/v2"

	"proto-gin-web/internal/auth"
	"proto-gin-web/internal/data"
	"proto-gin-web/internal/platform"
	appdb "proto-gin-web/internal/repo/pg"
)

// NewRouter wires middleware, views, and routes.
func NewRouter(cfg platform.Config) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Templates & static
	r.LoadHTMLGlob("internal/http/views/**/*")
	r.Static("/static", "web/static")

	// health
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// home
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Title":           "Proto — Gin + SSR",
			"Env":             cfg.Env,
			"BaseURL":         cfg.BaseURL,
			"SiteName":        cfg.SiteName,
			"SiteDescription": cfg.SiteDescription,
			"DocsURL":         "https://gin-gonic.com/docs/",
		})
	})

	// robots.txt
	r.GET("/robots.txt", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusOK, "User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", cfg.BaseURL)
	})

	// sitemap.xml (from DB posts)
	r.GET("/sitemap.xml", func(c *gin.Context) {
		pool, err := appdb.NewPool(c, cfg)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer pool.Close()
		q := data.New(pool)
		rows, err := q.ListPublishedPosts(c, 100, 0)
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

	// rss.xml (from DB posts)
	r.GET("/rss.xml", func(c *gin.Context) {
		pool, err := appdb.NewPool(c, cfg)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer pool.Close()
		q := data.New(pool)
		rows, err := q.ListPublishedPosts(c, 20, 0)
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

	// example API using sqlc
	r.GET("/api/articles", func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "10")
		offsetStr := c.DefaultQuery("offset", "0")
		limit, _ := strconv.ParseInt(limitStr, 10, 32)
		offset, _ := strconv.ParseInt(offsetStr, 10, 32)

		pool, err := appdb.NewPool(c, cfg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer pool.Close()

		q := data.New(pool)
		rows, err := q.ListArticles(c, int32(limit), int32(offset))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, rows)
	})

	// posts APIs with pagination + category/tag filter + sort
	r.GET("/api/posts", func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "10")
		offsetStr := c.DefaultQuery("offset", "0")
		limit, _ := strconv.ParseInt(limitStr, 10, 32)
		offset, _ := strconv.ParseInt(offsetStr, 10, 32)
		category := c.Query("category")
		tag := c.Query("tag")
		sort := c.DefaultQuery("sort", "created_at_desc")

		pool, err := appdb.NewPool(c, cfg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer pool.Close()

		q := data.New(pool)
		var rows []data.Post
		if category != "" {
			rows, err = q.ListPublishedPostsByCategorySorted(c, category, sort, int32(limit), int32(offset))
		} else if tag != "" {
			rows, err = q.ListPublishedPostsByTagSorted(c, tag, sort, int32(limit), int32(offset))
		} else {
			rows, err = q.ListPublishedPostsSorted(c, sort, int32(limit), int32(offset))
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, rows)
	})

	// SSR list with pagination (basic)
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

		pool, err := appdb.NewPool(c, cfg)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer pool.Close()

		q := data.New(pool)
		category := c.Query("category")
		tag := c.Query("tag")
		sort := c.DefaultQuery("sort", "created_at_desc")
		var rows []data.Post
		if category != "" {
			rows, err = q.ListPublishedPostsByCategorySorted(c, category, sort, int32(size), int32(offset))
		} else if tag != "" {
			rows, err = q.ListPublishedPostsByTagSorted(c, tag, sort, int32(size), int32(offset))
		} else {
			rows, err = q.ListPublishedPostsSorted(c, sort, int32(size), int32(offset))
		}
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Title":           "Posts — " + cfg.SiteName,
			"Env":             cfg.Env,
			"BaseURL":         cfg.BaseURL,
			"SiteName":        cfg.SiteName,
			"SiteDescription": cfg.SiteDescription,
			"Posts":           rows,
			"Page":            page,
			"Size":            size,
		})
	})

	// SSR post by slug (safe markdown rendering)
	r.GET("/posts/:slug", func(c *gin.Context) {
		slug := c.Param("slug")

		pool, err := appdb.NewPool(c, cfg)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer pool.Close()

		q := data.New(pool)
		p, err := q.GetPostBySlug(c, slug)
		if err != nil {
			c.String(http.StatusNotFound, "post not found")
			return
		}

		// Render Markdown -> HTML then sanitize
		unsafe := bf.Run([]byte(p.ContentMd))
		safe := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

		cats, _ := q.ListCategoriesByPostSlug(c, slug)
		tags, _ := q.ListTagsByPostSlug(c, slug)

		c.HTML(http.StatusOK, "post.tmpl", gin.H{
			"Title":           p.Title,
			"Summary":         p.Summary,
			"CoverURL":        p.CoverUrl,
			"ContentHTML":     template.HTML(string(safe)),
			"Categories":      cats,
			"Tags":            tags,
			"Env":             cfg.Env,
			"BaseURL":         cfg.BaseURL,
			"SiteName":        cfg.SiteName,
			"SiteDescription": cfg.SiteDescription,
		})
	})

	// Admin stub routes (basic, non-persistent session)
	r.GET("/admin/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "login with ?u=&p="})
	})
	r.POST("/admin/login", func(c *gin.Context) {
		u := c.PostForm("u")
		p := c.PostForm("p")
		if u == cfg.AdminUser && p == cfg.AdminPass {
			c.SetCookie("admin", "1", 3600, "/", "", false, true)
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false})
	})
	r.POST("/admin/logout", func(c *gin.Context) {
		c.SetCookie("admin", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	admin := r.Group("/admin", auth.AdminRequired())
	{
		// Admin CRUD for posts
		admin.POST("/posts", func(c *gin.Context) {
			var body struct {
				Title     string `json:"title" binding:"required"`
				Slug      string `json:"slug" binding:"required"`
				Summary   string `json:"summary"`
				ContentMD string `json:"content_md" binding:"required"`
				CoverURL  string `json:"cover_url"`
				Status    string `json:"status"`
				AuthorID  int64  `json:"author_id"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			pool, err := appdb.NewPool(c, cfg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer pool.Close()
			q := data.New(pool)
			row, err := q.CreatePost(c, data.CreatePostParams{
				Title:       body.Title,
				Slug:        body.Slug,
				Summary:     body.Summary,
				ContentMd:   body.ContentMD,
				CoverUrl:    body.CoverURL,
				Status:      body.Status,
				AuthorID:    body.AuthorID,
				PublishedAt: nil,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, row)
		})

		admin.PUT("/posts/:slug", func(c *gin.Context) {
			slug := c.Param("slug")
			var body struct {
				Title     string `json:"title" binding:"required"`
				Summary   string `json:"summary"`
				ContentMD string `json:"content_md" binding:"required"`
				CoverURL  string `json:"cover_url"`
				Status    string `json:"status"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			pool, err := appdb.NewPool(c, cfg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer pool.Close()
			q := data.New(pool)
			row, err := q.UpdatePostBySlug(c, data.UpdatePostBySlugParams{
				Slug:      slug,
				Title:     body.Title,
				Summary:   body.Summary,
				ContentMd: body.ContentMD,
				CoverUrl:  body.CoverURL,
				Status:    body.Status,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, row)
		})

		admin.DELETE("/posts/:slug", func(c *gin.Context) {
			slug := c.Param("slug")
			pool, err := appdb.NewPool(c, cfg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer pool.Close()
			q := data.New(pool)
			if err := q.DeletePostBySlug(c, slug); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		// Link/unlink categories to a post by slug
		admin.POST("/posts/:slug/categories/:cat", func(c *gin.Context) {
			pool, err := appdb.NewPool(c, cfg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer pool.Close()
			q := data.New(pool)
			if err := q.AddCategoryToPost(c, c.Param("slug"), c.Param("cat")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})
		admin.DELETE("/posts/:slug/categories/:cat", func(c *gin.Context) {
			pool, err := appdb.NewPool(c, cfg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer pool.Close()
			q := data.New(pool)
			if err := q.RemoveCategoryFromPost(c, c.Param("slug"), c.Param("cat")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		// Link/unlink tags to a post by slug
		admin.POST("/posts/:slug/tags/:tag", func(c *gin.Context) {
			pool, err := appdb.NewPool(c, cfg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer pool.Close()
			q := data.New(pool)
			if err := q.AddTagToPost(c, c.Param("slug"), c.Param("tag")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})
		admin.DELETE("/posts/:slug/tags/:tag", func(c *gin.Context) {
			pool, err := appdb.NewPool(c, cfg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer pool.Close()
			q := data.New(pool)
			if err := q.RemoveTagFromPost(c, c.Param("slug"), c.Param("tag")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		// Category CRUD
		admin.POST("/categories", func(c *gin.Context) {
			var body struct{ Name, Slug string }
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			pool, err := appdb.NewPool(c, cfg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer pool.Close()
			q := data.New(pool)
			row, err := q.CreateCategory(c, data.CreateCategoryParams{Name: body.Name, Slug: body.Slug})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, row)
		})
		admin.DELETE("/categories/:slug", func(c *gin.Context) {
			pool, err := appdb.NewPool(c, cfg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer pool.Close()
			q := data.New(pool)
			if err := q.DeleteCategoryBySlug(c, c.Param("slug")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		// Tag CRUD
		admin.POST("/tags", func(c *gin.Context) {
			var body struct{ Name, Slug string }
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			pool, err := appdb.NewPool(c, cfg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer pool.Close()
			q := data.New(pool)
			row, err := q.CreateTag(c, data.CreateTagParams{Name: body.Name, Slug: body.Slug})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, row)
		})
		admin.DELETE("/tags/:slug", func(c *gin.Context) {
			pool, err := appdb.NewPool(c, cfg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer pool.Close()
			q := data.New(pool)
			if err := q.DeleteTagBySlug(c, c.Param("slug")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})
	}

	return r
}
