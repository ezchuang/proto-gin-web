package adminuihttp

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	authdomain "proto-gin-web/internal/admin/auth/domain"
	adminview "proto-gin-web/internal/admin/ui/adapters/view"
	adminuisvc "proto-gin-web/internal/admin/ui/app"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/auth"
)

// RegisterUIRoutes mounts legacy SSR pages that still live inside the admin context.
func RegisterUIRoutes(r *gin.Engine, cfg platform.Config, adminSvc authdomain.AdminService, svc *adminuisvc.Service) {
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
			profile, err := currentAdminProfile(c, adminSvc)
			if err != nil {
				handleAdminProfileError(c, err)
				return
			}

			coverURL, err := resolveCoverInput(c)
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}

			title := c.PostForm("title")
			slug := c.PostForm("slug")
			summary := c.PostForm("summary")
			content := c.PostForm("content_md")
			status := c.DefaultPostForm("status", "draft")

			params := adminuisvc.CreatePostParams{
				Title:     title,
				Slug:      slug,
				Summary:   summary,
				ContentMD: content,
				CoverURL:  coverURL,
				Status:    status,
				AuthorID:  profile.ID,
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
			coverURL, err := resolveCoverInput(c)
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			params := adminuisvc.UpdatePostParams{
				Slug:      c.Param("slug"),
				Title:     c.PostForm("title"),
				Summary:   c.PostForm("summary"),
				ContentMD: c.PostForm("content_md"),
				CoverURL:  coverURL,
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

var errMissingAdminEmail = errors.New("admin session missing email")

func currentAdminProfile(c *gin.Context, adminSvc authdomain.AdminService) (authdomain.Admin, error) {
	email, err := c.Cookie("admin_email")
	if err != nil {
		return authdomain.Admin{}, errMissingAdminEmail
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return authdomain.Admin{}, errMissingAdminEmail
	}
	return adminSvc.GetProfile(c.Request.Context(), email)
}

func handleAdminProfileError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "failed to resolve admin session"
	switch {
	case errors.Is(err, errMissingAdminEmail):
		status = http.StatusUnauthorized
		message = "unauthorized"
	case errors.Is(err, authdomain.ErrAdminNotFound):
		status = http.StatusUnauthorized
		message = "admin profile not found"
	}
	c.String(status, message)
}

const (
	coverUploadDir      = "web/static/uploads"
	maxCoverUploadBytes = 5 << 20
)

var allowedCoverExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

func resolveCoverInput(c *gin.Context) (string, error) {
	file, err := c.FormFile("cover_file")
	if err == nil && file != nil && file.Size > 0 {
		if file.Size > maxCoverUploadBytes {
			return "", fmt.Errorf("cover file exceeds %d MB", maxCoverUploadBytes>>20)
		}
		return saveCoverUpload(c, file)
	}
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return "", err
	}
	return strings.TrimSpace(c.PostForm("cover_url")), nil
}

func saveCoverUpload(c *gin.Context, file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedCoverExt[ext] {
		return "", fmt.Errorf("unsupported cover format: %s", ext)
	}
	if err := os.MkdirAll(coverUploadDir, 0o755); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	destPath := filepath.Join(coverUploadDir, filename)
	if err := c.SaveUploadedFile(file, destPath); err != nil {
		return "", err
	}
	return "/static/uploads/" + filename, nil
}
