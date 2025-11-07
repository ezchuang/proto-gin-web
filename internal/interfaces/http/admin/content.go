package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/interfaces/http/view/responder"
)

func registerContentRoutes(group *gin.RouterGroup, postSvc domain.PostService, taxonomySvc domain.TaxonomyService) {
	group.POST("/posts", func(c *gin.Context) {
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
			responder.JSONError(c, http.StatusBadRequest, err.Error())
			return
		}

		ctx := c.Request.Context()
		cover := body.CoverURL
		input := domain.CreatePostInput{
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
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		responder.JSONSuccess(c, http.StatusOK, row)
	})

	group.PUT("/posts/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		var body struct {
			Title     string `json:"title" binding:"required"`
			Summary   string `json:"summary"`
			ContentMD string `json:"content_md" binding:"required"`
			CoverURL  string `json:"cover_url"`
			Status    string `json:"status"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			responder.JSONError(c, http.StatusBadRequest, err.Error())
			return
		}
		cover := body.CoverURL
		input := domain.UpdatePostInput{
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
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		responder.JSONSuccess(c, http.StatusOK, row)
	})

	group.DELETE("/posts/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		ctx := c.Request.Context()
		if err := postSvc.Delete(ctx, slug); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	})

	group.POST("/posts/:slug/categories/:cat", func(c *gin.Context) {
		ctx := c.Request.Context()
		if err := postSvc.AddCategory(ctx, c.Param("slug"), c.Param("cat")); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	})

	group.DELETE("/posts/:slug/categories/:cat", func(c *gin.Context) {
		ctx := c.Request.Context()
		if err := postSvc.RemoveCategory(ctx, c.Param("slug"), c.Param("cat")); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	})

	group.POST("/posts/:slug/tags/:tag", func(c *gin.Context) {
		ctx := c.Request.Context()
		if err := postSvc.AddTag(ctx, c.Param("slug"), c.Param("tag")); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	})

	group.DELETE("/posts/:slug/tags/:tag", func(c *gin.Context) {
		ctx := c.Request.Context()
		if err := postSvc.RemoveTag(ctx, c.Param("slug"), c.Param("tag")); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	})

	group.POST("/categories", func(c *gin.Context) {
		var body struct{ Name, Slug string }
		if err := c.ShouldBindJSON(&body); err != nil {
			responder.JSONError(c, http.StatusBadRequest, err.Error())
			return
		}
		ctx := c.Request.Context()
		category, err := taxonomySvc.CreateCategory(ctx, domain.CreateCategoryInput{
			Name: body.Name,
			Slug: body.Slug,
		})
		if err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		responder.JSONSuccess(c, http.StatusOK, category)
	})

	group.DELETE("/categories/:slug", func(c *gin.Context) {
		ctx := c.Request.Context()
		if err := taxonomySvc.DeleteCategory(ctx, c.Param("slug")); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	})

	group.POST("/tags", func(c *gin.Context) {
		var body struct{ Name, Slug string }
		if err := c.ShouldBindJSON(&body); err != nil {
			responder.JSONError(c, http.StatusBadRequest, err.Error())
			return
		}
		ctx := c.Request.Context()
		tag, err := taxonomySvc.CreateTag(ctx, domain.CreateTagInput{
			Name: body.Name,
			Slug: body.Slug,
		})
		if err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		responder.JSONSuccess(c, http.StatusOK, tag)
	})

	group.DELETE("/tags/:slug", func(c *gin.Context) {
		ctx := c.Request.Context()
		if err := taxonomySvc.DeleteTag(ctx, c.Param("slug")); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	})
}
