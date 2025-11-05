package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/auth"
)

// registerAdminUIRoutes mounts simple SSR pages for admin management (MVP).
func registerAdminUIRoutes(r *gin.Engine, cfg platform.Config, postSvc domain.PostService) {
	// Redirect root to posts list
	r.GET("/admin/ui", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/admin/ui/posts")
	})

	admin := r.Group("/admin/ui", auth.AdminRequired())
	{
		admin.GET("/posts", func(c *gin.Context) {
			ctx := c.Request.Context()
			rows, err := postSvc.ListPublished(ctx, domain.ListPostsOptions{Limit: 50})
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			renderHTML(c, http.StatusOK, "admin_posts.tmpl", gin.H{
				"Title":           "Admin · Posts · " + cfg.SiteName,
				"Env":             cfg.Env,
				"BaseURL":         cfg.BaseURL,
				"SiteName":        cfg.SiteName,
				"SiteDescription": cfg.SiteDescription,
				"Posts":           rows,
			})
		})

		admin.GET("/posts/new", func(c *gin.Context) {
			renderHTML(c, http.StatusOK, "admin_post_form.tmpl", gin.H{
				"Title":           "Admin · New Post · " + cfg.SiteName,
				"Env":             cfg.Env,
				"BaseURL":         cfg.BaseURL,
				"SiteName":        cfg.SiteName,
				"SiteDescription": cfg.SiteDescription,
				"IsNew":           true,
			})
		})

		admin.POST("/posts/new", func(c *gin.Context) {
			title := c.PostForm("title")
			slug := c.PostForm("slug")
			summary := c.PostForm("summary")
			content := c.PostForm("content_md")
			cover := c.PostForm("cover_url")
			status := c.DefaultPostForm("status", "draft")
			authorStr := c.DefaultPostForm("author_id", "1")
			authorID, _ := strconv.ParseInt(authorStr, 10, 64)
			var publishedAt *time.Time
			_ = publishedAt // keep nil for now
			input := domain.CreatePostInput{
				Title:       title,
				Slug:        slug,
				Summary:     summary,
				ContentMD:   content,
				Status:      status,
				AuthorID:    authorID,
				PublishedAt: publishedAt,
			}
			if cover != "" {
				input.CoverURL = &cover
			}
			if _, err := postSvc.Create(c.Request.Context(), input); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts")
		})

		admin.GET("/posts/:slug/edit", func(c *gin.Context) {
			slug := c.Param("slug")
			result, err := postSvc.GetBySlug(c.Request.Context(), slug)
			if err != nil {
				c.String(http.StatusNotFound, "post not found")
				return
			}
			renderHTML(c, http.StatusOK, "admin_post_form.tmpl", gin.H{
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
		})

		admin.POST("/posts/:slug", func(c *gin.Context) {
			slug := c.Param("slug")
			title := c.PostForm("title")
			summary := c.PostForm("summary")
			content := c.PostForm("content_md")
			cover := c.PostForm("cover_url")
			status := c.DefaultPostForm("status", "draft")
			input := domain.UpdatePostInput{
				Slug:      slug,
				Title:     title,
				Summary:   summary,
				ContentMD: content,
				Status:    status,
			}
			if cover != "" {
				input.CoverURL = &cover
			}
			if _, err := postSvc.Update(c.Request.Context(), input); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts/"+slug+"/edit")
		})

		admin.POST("/posts/:slug/delete", func(c *gin.Context) {
			slug := c.Param("slug")
			if err := postSvc.Delete(c.Request.Context(), slug); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts")
		})

		admin.POST("/posts/:slug/categories/add", func(c *gin.Context) {
			slug := c.Param("slug")
			cat := c.PostForm("category_slug")
			if err := postSvc.AddCategory(c.Request.Context(), slug, cat); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts/"+slug+"/edit")
		})
		admin.POST("/posts/:slug/categories/remove", func(c *gin.Context) {
			slug := c.Param("slug")
			cat := c.PostForm("category_slug")
			if err := postSvc.RemoveCategory(c.Request.Context(), slug, cat); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts/"+slug+"/edit")
		})

		admin.POST("/posts/:slug/tags/add", func(c *gin.Context) {
			slug := c.Param("slug")
			tag := c.PostForm("tag_slug")
			if err := postSvc.AddTag(c.Request.Context(), slug, tag); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts/"+slug+"/edit")
		})
		admin.POST("/posts/:slug/tags/remove", func(c *gin.Context) {
			slug := c.Param("slug")
			tag := c.PostForm("tag_slug")
			if err := postSvc.RemoveTag(c.Request.Context(), slug, tag); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts/"+slug+"/edit")
		})
	}
}
