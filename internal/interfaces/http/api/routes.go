package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
)

// RegisterRoutes attaches JSON endpoints for posts.
func RegisterRoutes(r *gin.Engine, postSvc domain.PostService) {
	api := r.Group("/api")
	{
		api.GET("/posts", func(c *gin.Context) {
			limit := int32(10)
			offset := int32(0)
			if v := c.DefaultQuery("limit", "10"); v != "" {
				if parsed, err := strconv.ParseInt(v, 10, 32); err == nil {
					limit = int32(parsed)
				}
			}
			if v := c.DefaultQuery("offset", "0"); v != "" {
				if parsed, err := strconv.ParseInt(v, 10, 32); err == nil {
					offset = int32(parsed)
				}
			}
			rows, err := postSvc.ListPublished(c.Request.Context(), domain.ListPostsOptions{
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, rows)
		})

		api.GET("/posts/:slug", func(c *gin.Context) {
			row, err := postSvc.GetBySlug(c.Request.Context(), c.Param("slug"))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
				return
			}
			c.JSON(http.StatusOK, row)
		})
	}
}
