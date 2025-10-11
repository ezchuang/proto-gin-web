package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/core"
	appdb "proto-gin-web/internal/repo/pg"
)

// registerAPIRoutes attaches JSON endpoints for articles and posts.
func registerAPIRoutes(r *gin.Engine, postSvc core.PostService, queries *appdb.Queries) {
	r.GET("/api/articles", func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "10")
		offsetStr := c.DefaultQuery("offset", "0")
		limit, _ := strconv.ParseInt(limitStr, 10, 32)
		offset, _ := strconv.ParseInt(offsetStr, 10, 32)

		ctx := c.Request.Context()
		rows, err := queries.ListArticles(ctx, int32(limit), int32(offset))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, rows)
	})

	r.GET("/api/posts", func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "10")
		offsetStr := c.DefaultQuery("offset", "0")
		limit, _ := strconv.ParseInt(limitStr, 10, 32)
		offset, _ := strconv.ParseInt(offsetStr, 10, 32)
		category := c.Query("category")
		tag := c.Query("tag")
		sort := c.DefaultQuery("sort", "created_at_desc")

		ctx := c.Request.Context()
		rows, err := postSvc.ListPublished(ctx, core.ListPostsOptions{
			Category: category,
			Tag:      tag,
			Sort:     sort,
			Limit:    int32(limit),
			Offset:   int32(offset),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, rows)
	})
}
