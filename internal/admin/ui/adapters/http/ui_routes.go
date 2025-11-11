package adminuihttp

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	adminview "proto-gin-web/internal/admin/ui/adapters/view"
	adminuisvc "proto-gin-web/internal/admin/ui/app"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/auth"
)

// RegisterUIRoutes mounts legacy SSR pages that still live inside the admin context.
func RegisterUIRoutes(r *gin.Engine, cfg platform.Config, svc *adminuisvc.Service) {
	// Redirect root to posts list
	r.GET("/admin/ui", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/admin/ui/posts")
	})

	admin := r.Group("/admin/ui", auth.AdminRequired())
	{
		admin.GET("/posts", func(c *gin.Context) {
			ctx := c.Request.Context()
			rows, err := svc.ListPosts(ctx, 50)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			adminview.AdminPostsPage(c, cfg, rows)
		})

		admin.GET("/posts/new", func(c *gin.Context) {
			adminview.AdminPostFormNew(c, cfg)
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
			params := adminuisvc.CreatePostParams{
				Title:     title,
				Slug:      slug,
				Summary:   summary,
				ContentMD: content,
				CoverURL:  cover,
				Status:    status,
				AuthorID:  authorID,
			}
			if _, err := svc.CreatePost(c.Request.Context(), params); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts")
		})

		admin.GET("/posts/:slug/edit", func(c *gin.Context) {
			slug := c.Param("slug")
			result, err := svc.GetPost(c.Request.Context(), slug)
			if err != nil {
				c.String(http.StatusNotFound, "post not found")
				return
			}
			adminview.AdminPostFormEdit(c, cfg, result)
		})

		admin.POST("/posts/:slug", func(c *gin.Context) {
			params := adminuisvc.UpdatePostParams{
				Slug:      c.Param("slug"),
				Title:     c.PostForm("title"),
				Summary:   c.PostForm("summary"),
				ContentMD: c.PostForm("content_md"),
				CoverURL:  c.PostForm("cover_url"),
				Status:    c.DefaultPostForm("status", "draft"),
			}
			if _, err := svc.UpdatePost(c.Request.Context(), params); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts/"+params.Slug+"/edit")
		})

		admin.POST("/posts/:slug/delete", func(c *gin.Context) {
			if err := svc.DeletePost(c.Request.Context(), c.Param("slug")); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts")
		})

		admin.POST("/posts/:slug/categories/add", func(c *gin.Context) {
			slug := c.Param("slug")
			cat := c.PostForm("category_slug")
			if err := svc.AddCategory(c.Request.Context(), slug, cat); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts/"+slug+"/edit")
		})
		admin.POST("/posts/:slug/categories/remove", func(c *gin.Context) {
			slug := c.Param("slug")
			cat := c.PostForm("category_slug")
			if err := svc.RemoveCategory(c.Request.Context(), slug, cat); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts/"+slug+"/edit")
		})

		admin.POST("/posts/:slug/tags/add", func(c *gin.Context) {
			slug := c.Param("slug")
			tag := c.PostForm("tag_slug")
			if err := svc.AddTag(c.Request.Context(), slug, tag); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts/"+slug+"/edit")
		})
		admin.POST("/posts/:slug/tags/remove", func(c *gin.Context) {
			slug := c.Param("slug")
			tag := c.PostForm("tag_slug")
			if err := svc.RemoveTag(c.Request.Context(), slug, tag); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			c.Redirect(http.StatusFound, "/admin/ui/posts/"+slug+"/edit")
		})
	}
}
