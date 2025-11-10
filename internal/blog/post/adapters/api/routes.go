package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/interfaces/http/view/presenter"
	"proto-gin-web/internal/interfaces/http/view/responder"
)

// RegisterRoutes attaches JSON endpoints for posts.
func RegisterRoutes(r *gin.Engine, postSvc domain.PostService) {
	api := r.Group("/api")
	{
		api.GET("/posts", listPostsHandler(postSvc))
		api.GET("/posts/:slug", getPostHandler(postSvc))
	}
}

// listPostsHandler godoc
// @Summary      List published posts
// @Description  Retrieves a paginated list of published posts.
// @Tags         Public
// @Produce      json
// @Param        limit   query     int  false  "Number of posts to return" default(10)
// @Param        offset  query     int  false  "Pagination offset" default(0)
// @Success      200  {object}  postListResponse
// @Failure      500  {object}  errorResponse
// @Router       /api/posts [get]
func listPostsHandler(postSvc domain.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		responder.JSONSuccess(c, http.StatusOK, presenter.BuildPublicPosts(rows))
	}
}

// getPostHandler godoc
// @Summary      Get a post by slug
// @Description  Retrieves a post together with its categories and tags.
// @Tags         Public
// @Produce      json
// @Param        slug  path      string  true  "Post slug"
// @Success      200  {object}  postResponse
// @Failure      404  {object}  errorResponse
// @Router       /api/posts/{slug} [get]
func getPostHandler(postSvc domain.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		row, err := postSvc.GetBySlug(c.Request.Context(), c.Param("slug"))
		if err != nil {
			responder.JSONError(c, http.StatusNotFound, "post not found")
			return
		}
		responder.JSONSuccess(c, http.StatusOK, presenter.BuildPublicPostWithRelations(row))
	}
}
