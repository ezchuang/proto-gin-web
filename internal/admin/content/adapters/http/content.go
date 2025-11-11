package contenthttp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	admincontentusecase "proto-gin-web/internal/admin/content/app"
	postdomain "proto-gin-web/internal/blog/post/domain"
	taxdomain "proto-gin-web/internal/blog/taxonomy/domain"
	"proto-gin-web/internal/platform/http/responder"
)

// AdminCreatePostRequest describes the payload to create a post.
type AdminCreatePostRequest struct {
	Title     string `json:"title" binding:"required"`
	Slug      string `json:"slug" binding:"required"`
	Summary   string `json:"summary"`
	ContentMD string `json:"content_md" binding:"required"`
	CoverURL  string `json:"cover_url"`
	Status    string `json:"status"`
	AuthorID  int64  `json:"author_id"`
}

// AdminUpdatePostRequest describes the payload to update a post.
type AdminUpdatePostRequest struct {
	Title     string `json:"title" binding:"required"`
	Summary   string `json:"summary"`
	ContentMD string `json:"content_md" binding:"required"`
	CoverURL  string `json:"cover_url"`
	Status    string `json:"status"`
}

// AdminTaxonomyRequest describes a category/tag payload.
type AdminTaxonomyRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

func RegisterRoutes(group *gin.RouterGroup, contentSvc *admincontentusecase.Service) {
	group.POST("/posts", createPostHandler(contentSvc))
	group.PUT("/posts/:slug", updatePostHandler(contentSvc))
	group.DELETE("/posts/:slug", deletePostHandler(contentSvc))
	group.POST("/posts/:slug/categories/:cat", addCategoryHandler(contentSvc))
	group.DELETE("/posts/:slug/categories/:cat", removeCategoryHandler(contentSvc))
	group.POST("/posts/:slug/tags/:tag", addTagHandler(contentSvc))
	group.DELETE("/posts/:slug/tags/:tag", removeTagHandler(contentSvc))
	group.POST("/categories", createCategoryHandler(contentSvc))
	group.DELETE("/categories/:slug", deleteCategoryHandler(contentSvc))
	group.POST("/tags", createTagHandler(contentSvc))
	group.DELETE("/tags/:slug", deleteTagHandler(contentSvc))
}

// createPostHandler godoc
// @Summary      Create a post
// @Description  Creates a post for the admin UI.
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        payload  body      AdminCreatePostRequest  true  "Post payload"
// @Success      200      {object}  admincontentusecase.AdminPostResponse
// @Failure      400      {object}  admincontentusecase.AdminErrorResponse
// @Failure      500      {object}  admincontentusecase.AdminErrorResponse
// @Router       /admin/posts [post]
func createPostHandler(contentSvc *admincontentusecase.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body AdminCreatePostRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			responder.JSONError(c, http.StatusBadRequest, err.Error())
			return
		}

		cover := body.CoverURL
		input := postdomain.CreatePostInput{
			Title:       body.Title,
			Slug:        body.Slug,
			Summary:     body.Summary,
			ContentMD:   body.ContentMD,
			Status:      body.Status,
			AuthorID:    body.AuthorID,
			PublishedAt: nil,
		}
		input.CoverURL = &cover

		row, err := contentSvc.CreatePost(c.Request.Context(), input)
		if err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		responder.JSONSuccess(c, http.StatusOK, row)
	}
}

// updatePostHandler godoc
// @Summary      Update a post
// @Description  Updates a post identified by slug.
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        slug     path      string                   true  "Post slug"
// @Param        payload  body      AdminUpdatePostRequest   true  "Post payload"
// @Success      200      {object}  admincontentusecase.AdminPostResponse
// @Failure      400      {object}  admincontentusecase.AdminErrorResponse
// @Failure      500      {object}  admincontentusecase.AdminErrorResponse
// @Router       /admin/posts/{slug} [put]
func updatePostHandler(contentSvc *admincontentusecase.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		var body AdminUpdatePostRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			responder.JSONError(c, http.StatusBadRequest, err.Error())
			return
		}
		cover := body.CoverURL
		input := postdomain.UpdatePostInput{
			Slug:      slug,
			Title:     body.Title,
			Summary:   body.Summary,
			ContentMD: body.ContentMD,
			Status:    body.Status,
		}
		input.CoverURL = &cover

		row, err := contentSvc.UpdatePost(c.Request.Context(), input)
		if err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		responder.JSONSuccess(c, http.StatusOK, row)
	}
}

// deletePostHandler godoc
// @Summary      Delete a post
// @Description  Deletes a post identified by slug.
// @Tags         Admin
// @Param        slug  path  string  true  "Post slug"
// @Success      204  {string}  string  ""
// @Failure      500  {object}  admincontentusecase.AdminErrorResponse
// @Router       /admin/posts/{slug} [delete]
func deletePostHandler(contentSvc *admincontentusecase.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if err := contentSvc.DeletePost(c.Request.Context(), slug); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// addCategoryHandler godoc
// @Summary      Attach a category to a post
// @Tags         Admin
// @Param        slug  path  string  true  "Post slug"
// @Param        cat   path  string  true  "Category slug"
// @Success      204 {string} string ""
// @Failure      500 {object} admincontentusecase.AdminErrorResponse
// @Router       /admin/posts/{slug}/categories/{cat} [post]
func addCategoryHandler(contentSvc *admincontentusecase.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := contentSvc.AddCategory(c.Request.Context(), c.Param("slug"), c.Param("cat")); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// removeCategoryHandler godoc
// @Summary      Remove a category from a post
// @Tags         Admin
// @Param        slug  path  string  true  "Post slug"
// @Param        cat   path  string  true  "Category slug"
// @Success      204 {string} string ""
// @Failure      500 {object} admincontentusecase.AdminErrorResponse
// @Router       /admin/posts/{slug}/categories/{cat} [delete]
func removeCategoryHandler(contentSvc *admincontentusecase.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := contentSvc.RemoveCategory(c.Request.Context(), c.Param("slug"), c.Param("cat")); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// addTagHandler godoc
// @Summary      Attach a tag to a post
// @Tags         Admin
// @Param        slug  path  string  true  "Post slug"
// @Param        tag   path  string  true  "Tag slug"
// @Success      204 {string} string ""
// @Failure      500 {object} admincontentusecase.AdminErrorResponse
// @Router       /admin/posts/{slug}/tags/{tag} [post]
func addTagHandler(contentSvc *admincontentusecase.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := contentSvc.AddTag(c.Request.Context(), c.Param("slug"), c.Param("tag")); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// removeTagHandler godoc
// @Summary      Remove a tag from a post
// @Tags         Admin
// @Param        slug  path  string  true  "Post slug"
// @Param        tag   path  string  true  "Tag slug"
// @Success      204 {string} string ""
// @Failure      500 {object} admincontentusecase.AdminErrorResponse
// @Router       /admin/posts/{slug}/tags/{tag} [delete]
func removeTagHandler(contentSvc *admincontentusecase.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := contentSvc.RemoveTag(c.Request.Context(), c.Param("slug"), c.Param("tag")); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// createCategoryHandler godoc
// @Summary      Create category
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        payload  body      AdminTaxonomyRequest  true  "Category payload"
// @Success      200      {object}  admincontentusecase.AdminCategoryResponse
// @Failure      400      {object}  admincontentusecase.AdminErrorResponse
// @Failure      500      {object}  admincontentusecase.AdminErrorResponse
// @Router       /admin/categories [post]
func createCategoryHandler(contentSvc *admincontentusecase.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body AdminTaxonomyRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			responder.JSONError(c, http.StatusBadRequest, err.Error())
			return
		}
		category, err := contentSvc.CreateCategory(c.Request.Context(), taxdomain.CreateCategoryInput{
			Name: body.Name,
			Slug: body.Slug,
		})
		if err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		responder.JSONSuccess(c, http.StatusOK, category)
	}
}

// deleteCategoryHandler godoc
// @Summary      Delete category
// @Tags         Admin
// @Param        slug  path  string  true  "Category slug"
// @Success      204  {string} string ""
// @Failure      500  {object}  admincontentusecase.AdminErrorResponse
// @Router       /admin/categories/{slug} [delete]
func deleteCategoryHandler(contentSvc *admincontentusecase.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := contentSvc.DeleteCategory(c.Request.Context(), c.Param("slug")); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// createTagHandler godoc
// @Summary      Create tag
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        payload  body      AdminTaxonomyRequest  true  "Tag payload"
// @Success      200      {object}  admincontentusecase.AdminTagResponse
// @Failure      400      {object}  admincontentusecase.AdminErrorResponse
// @Failure      500      {object}  admincontentusecase.AdminErrorResponse
// @Router       /admin/tags [post]
func createTagHandler(contentSvc *admincontentusecase.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body AdminTaxonomyRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			responder.JSONError(c, http.StatusBadRequest, err.Error())
			return
		}
		tag, err := contentSvc.CreateTag(c.Request.Context(), taxdomain.CreateTagInput{
			Name: body.Name,
			Slug: body.Slug,
		})
		if err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		responder.JSONSuccess(c, http.StatusOK, tag)
	}
}

// deleteTagHandler godoc
// @Summary      Delete tag
// @Tags         Admin
// @Param        slug  path  string  true  "Tag slug"
// @Success      204  {string} string ""
// @Failure      500  {object}  admincontentusecase.AdminErrorResponse
// @Router       /admin/tags/{slug} [delete]
func deleteTagHandler(contentSvc *admincontentusecase.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := contentSvc.DeleteTag(c.Request.Context(), c.Param("slug")); err != nil {
			responder.JSONError(c, http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusNoContent)
	}
}
