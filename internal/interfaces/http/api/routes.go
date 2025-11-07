package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/interfaces/http/view/responder"
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
				responder.JSONError(c, http.StatusInternalServerError, err.Error())
				return
			}
			responder.JSONSuccess(c, http.StatusOK, rows)
		})

		api.GET("/posts/:slug", func(c *gin.Context) {
			row, err := postSvc.GetBySlug(c.Request.Context(), c.Param("slug"))
			if err != nil {
				responder.JSONError(c, http.StatusNotFound, "post not found")
				return
			}
			responder.JSONSuccess(c, http.StatusOK, row)
		})
	}
}
