package adminuihttp

import (
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	authdomain "proto-gin-web/internal/admin/auth/domain"
	authsession "proto-gin-web/internal/admin/auth/session"
	adminview "proto-gin-web/internal/admin/ui/adapters/view"
	adminuisvc "proto-gin-web/internal/admin/ui/app"
	adminusecase "proto-gin-web/internal/application/admin"
	"proto-gin-web/internal/platform/config"
)

// RegisterUIRoutes mounts legacy SSR pages that still live inside the admin context.
func RegisterUIRoutes(r *gin.Engine, cfg config.Config, adminSvc adminusecase.AdminService, svc *adminuisvc.Service, sessionMgr *authsession.Manager, sessionGuard gin.HandlerFunc) {
	// Redirect root to posts list
	r.GET("/admin/ui", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/admin/ui/posts")
	})

	admin := r.Group("/admin/ui", sessionGuard)
	{
		admin.GET("/posts", func(c *gin.Context) {
			ctx := c.Request.Context()
			rows, err := svc.ListPosts(ctx, 50)
			if err != nil {
				logAdminUIError(c, "list posts", err)
				c.String(http.StatusInternalServerError, "internal server error")
				return
			}
			adminview.AdminPostsPage(c, cfg, rows)
		})

		admin.GET("/posts/new", func(c *gin.Context) {
			adminview.AdminPostFormNew(c, cfg)
		})

		admin.POST("/posts/new", func(c *gin.Context) {
			profile, ok := adminProfileFromContext(c)
			if !ok {
				handleAdminProfileError(c, errMissingAdminEmail)
				return
			}

			coverURL, err := resolveCoverInput(c)
			if err != nil {
				redirectWithError(c, "/admin/ui/posts/new", err.Error(), err)
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
				redirectWithError(c, "/admin/ui/posts/new", "failed to create post", err)
				return
			}
			redirectWithSuccess(c, "/admin/ui/posts", "post created")
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
				redirectWithError(c, "/admin/ui/posts/"+c.Param("slug")+"/edit", err.Error(), err)
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
				redirectWithError(c, "/admin/ui/posts/"+params.Slug+"/edit", "failed to update post", err)
				return
			}
			redirectWithSuccess(c, "/admin/ui/posts/"+params.Slug+"/edit", "post updated")
		})

		admin.POST("/posts/:slug/delete", func(c *gin.Context) {
			if err := svc.DeletePost(c.Request.Context(), c.Param("slug")); err != nil {
				redirectWithError(c, "/admin/ui/posts", "failed to delete post", err)
				return
			}
			redirectWithSuccess(c, "/admin/ui/posts", "post deleted")
		})

		admin.POST("/posts/:slug/categories/add", func(c *gin.Context) {
			slug := c.Param("slug")
			cat := c.PostForm("category_slug")
			if err := svc.AddCategory(c.Request.Context(), slug, cat); err != nil {
				redirectWithError(c, "/admin/ui/posts/"+slug+"/edit", "failed to add category", err)
				return
			}
			redirectWithSuccess(c, "/admin/ui/posts/"+slug+"/edit", "category added")
		})
		admin.POST("/posts/:slug/categories/remove", func(c *gin.Context) {
			slug := c.Param("slug")
			cat := c.PostForm("category_slug")
			if err := svc.RemoveCategory(c.Request.Context(), slug, cat); err != nil {
				redirectWithError(c, "/admin/ui/posts/"+slug+"/edit", "failed to remove category", err)
				return
			}
			redirectWithSuccess(c, "/admin/ui/posts/"+slug+"/edit", "category removed")
		})

		admin.POST("/posts/:slug/tags/add", func(c *gin.Context) {
			slug := c.Param("slug")
			tag := c.PostForm("tag_slug")
			if err := svc.AddTag(c.Request.Context(), slug, tag); err != nil {
				redirectWithError(c, "/admin/ui/posts/"+slug+"/edit", "failed to add tag", err)
				return
			}
			redirectWithSuccess(c, "/admin/ui/posts/"+slug+"/edit", "tag added")
		})
		admin.POST("/posts/:slug/tags/remove", func(c *gin.Context) {
			slug := c.Param("slug")
			tag := c.PostForm("tag_slug")
			if err := svc.RemoveTag(c.Request.Context(), slug, tag); err != nil {
				redirectWithError(c, "/admin/ui/posts/"+slug+"/edit", "failed to remove tag", err)
				return
			}
			redirectWithSuccess(c, "/admin/ui/posts/"+slug+"/edit", "tag removed")
		})
	}
}

var errMissingAdminEmail = errors.New("admin session missing email")

func adminProfileFromContext(c *gin.Context) (authdomain.Admin, bool) {
	if v, ok := c.Get("admin_profile"); ok {
		if profile, ok := v.(authdomain.Admin); ok {
			return profile, true
		}
	}
	return authdomain.Admin{}, false
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

func redirectWithError(c *gin.Context, path, message string, err error) {
	logAdminUIError(c, message, err)
	c.Redirect(http.StatusFound, addFlashQuery(path, "error", message))
}

func redirectWithSuccess(c *gin.Context, path, message string) {
	c.Redirect(http.StatusFound, addFlashQuery(path, "success", message))
}

func addFlashQuery(path, key, message string) string {
	if key == "" || message == "" {
		return path
	}
	u, err := url.Parse(path)
	if err != nil {
		return path
	}
	values := u.Query()
	values.Set(key, message)
	u.RawQuery = values.Encode()
	return u.String()
}

func logAdminUIError(c *gin.Context, context string, err error) {
	if err == nil {
		return
	}
	slog.Default().Error("admin ui: "+context,
		slog.String("path", c.FullPath()),
		slog.Any("err", err))
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
		logAdminUIError(c, "read cover upload", err)
		return "", errors.New("failed to read cover file")
	}
	return strings.TrimSpace(c.PostForm("cover_url")), nil
}

func saveCoverUpload(c *gin.Context, file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedCoverExt[ext] {
		return "", fmt.Errorf("unsupported cover format: %s", ext)
	}
	if err := os.MkdirAll(coverUploadDir, 0o755); err != nil {
		logAdminUIError(c, "create cover upload dir", err)
		return "", errors.New("failed to save cover file")
	}
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	destPath := filepath.Join(coverUploadDir, filename)
	if err := c.SaveUploadedFile(file, destPath); err != nil {
		logAdminUIError(c, "save cover upload", err)
		return "", errors.New("failed to save cover file")
	}
	return "/static/uploads/" + filename, nil
}

