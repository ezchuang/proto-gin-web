package public

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	bf "github.com/russross/blackfriday/v2"

	postview "proto-gin-web/internal/contexts/blog/post/adapters/view"
	postdomain "proto-gin-web/internal/contexts/blog/post/domain"
	postusecase "proto-gin-web/internal/contexts/blog/post/usecase"
	"proto-gin-web/internal/platform/config"
)

func registerContentRoutes(r *gin.Engine, cfg config.Config, postSvc postusecase.PostService) {
	r.GET("/", func(c *gin.Context) {
		postview.PublicLanding(c, cfg)
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
		rows, err := postSvc.ListPublished(ctx, postdomain.ListPostsOptions{
			Category: category,
			Tag:      tag,
			Sort:     sort,
			Limit:    int32(size),
			Offset:   int32(offset),
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "internal server error")
			return
		}
		postview.PublicPosts(c, cfg, rows, page, size)
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

		postview.PublicPostDetail(c, cfg, result, template.HTML(string(safe)))
	})
}




