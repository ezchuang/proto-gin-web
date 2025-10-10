package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/auth"
	"proto-gin-web/internal/core"
	"proto-gin-web/internal/platform"
	appdb "proto-gin-web/internal/repo/pg"
)

// registerAdminRoutes wires admin authentication and CRUD handlers.
func registerAdminRoutes(r *gin.Engine, cfg platform.Config, postSvc core.PostService, queries *appdb.Queries) {
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

			ctx := c.Request.Context()
			cover := body.CoverURL
			input := core.CreatePostInput{
				Title:       body.Title,
				Slug:        body.Slug,
				Summary:     body.Summary,
				ContentMD:   body.ContentMD,
				Status:      body.Status,
				AuthorID:    body.AuthorID,
				PublishedAt: nil,
			}
			input.CoverURL = &cover

			row, err := postSvc.Create(ctx, input)
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
			cover := body.CoverURL
			input := core.UpdatePostInput{
				Slug:      slug,
				Title:     body.Title,
				Summary:   body.Summary,
				ContentMD: body.ContentMD,
				Status:    body.Status,
			}
			input.CoverURL = &cover

			ctx := c.Request.Context()
			row, err := postSvc.Update(ctx, input)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, row)
		})

		admin.DELETE("/posts/:slug", func(c *gin.Context) {
			slug := c.Param("slug")
			ctx := c.Request.Context()
			if err := postSvc.Delete(ctx, slug); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		admin.POST("/posts/:slug/categories/:cat", func(c *gin.Context) {
			ctx := c.Request.Context()
			if err := postSvc.AddCategory(ctx, c.Param("slug"), c.Param("cat")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})
		admin.DELETE("/posts/:slug/categories/:cat", func(c *gin.Context) {
			ctx := c.Request.Context()
			if err := postSvc.RemoveCategory(ctx, c.Param("slug"), c.Param("cat")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		admin.POST("/posts/:slug/tags/:tag", func(c *gin.Context) {
			ctx := c.Request.Context()
			if err := postSvc.AddTag(ctx, c.Param("slug"), c.Param("tag")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})
		admin.DELETE("/posts/:slug/tags/:tag", func(c *gin.Context) {
			ctx := c.Request.Context()
			if err := postSvc.RemoveTag(ctx, c.Param("slug"), c.Param("tag")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		admin.POST("/categories", func(c *gin.Context) {
			var body struct{ Name, Slug string }
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx := c.Request.Context()
			row, err := queries.CreateCategory(ctx, appdb.CreateCategoryParams{Name: body.Name, Slug: body.Slug})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, row)
		})
		admin.DELETE("/categories/:slug", func(c *gin.Context) {
			ctx := c.Request.Context()
			if err := queries.DeleteCategoryBySlug(ctx, c.Param("slug")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		admin.POST("/tags", func(c *gin.Context) {
			var body struct{ Name, Slug string }
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx := c.Request.Context()
			row, err := queries.CreateTag(ctx, appdb.CreateTagParams{Name: body.Name, Slug: body.Slug})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, row)
		})
		admin.DELETE("/tags/:slug", func(c *gin.Context) {
			ctx := c.Request.Context()
			if err := queries.DeleteTagBySlug(ctx, c.Param("slug")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})
	}
}
