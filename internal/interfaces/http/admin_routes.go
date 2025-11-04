package http

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/argon2"

	"proto-gin-web/internal/domain"
	appdb "proto-gin-web/internal/infrastructure/pg"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/auth"
)

// registerAdminRoutes wires admin authentication and CRUD handlers.
func registerAdminRoutes(r *gin.Engine, cfg platform.Config, postSvc domain.PostService, queries *appdb.Queries) {
	loginLimiter := NewIPRateLimiter(5, time.Minute)
	registerLimiter := NewIPRateLimiter(3, time.Minute)

	r.GET("/admin/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin_login.tmpl", gin.H{
			"SiteName":        cfg.SiteName,
			"SiteDescription": cfg.SiteDescription,
			"Env":             cfg.Env,
			"BaseURL":         cfg.BaseURL,
			"Error":           c.Query("error"),
		})
	})
	r.POST("/admin/login", loginLimiter, func(c *gin.Context) {
		u := c.PostForm("u")
		p := c.PostForm("p")
		isForm := wantsHTMLResponse(c)

		ctx := c.Request.Context()
		userRecord, err := queries.GetUserByEmail(ctx, u)
		authenticated := false
		if err == nil {
			if ok, verifyErr := verifyArgon2idHash(userRecord.PasswordHash, p); verifyErr == nil && ok {
				authenticated = true
			} else if verifyErr != nil {
				slog.Error("argon2 verification failed",
					slog.String("user", u),
					slog.Any("err", verifyErr))
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			slog.Error("admin login lookup failed",
				slog.String("user", u),
				slog.String("ip", c.ClientIP()),
				slog.Any("err", err))
		}

		// Fallback to legacy config credentials if DB lookup failed
		if !authenticated && u == cfg.AdminUser && p == cfg.AdminPass {
			authenticated = true
		}

		if authenticated {
			secureCookie := cfg.Env == "production"
			c.SetSameSite(http.SameSiteStrictMode)
			c.SetCookie("admin", "1", 3600, "/", "", secureCookie, true)
			slog.Info("admin login succeeded",
				slog.String("user", u),
				slog.String("ip", c.ClientIP()))
			if isForm {
				c.Redirect(http.StatusFound, "/admin")
				return
			}
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}

		slog.Warn("admin login failed",
			slog.String("user", u),
			slog.String("ip", c.ClientIP()))
		if isForm {
			c.Redirect(http.StatusFound, "/admin/login?error=Invalid+credentials")
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false})
	})
	r.POST("/admin/logout", func(c *gin.Context) {
		secureCookie := cfg.Env == "production"
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("admin", "", -1, "/", "", secureCookie, true)
		if wantsHTMLResponse(c) {
			c.Redirect(http.StatusFound, "/admin/login")
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.GET("/admin/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin_register.tmpl", gin.H{
			"SiteName":        cfg.SiteName,
			"SiteDescription": cfg.SiteDescription,
			"Env":             cfg.Env,
			"BaseURL":         cfg.BaseURL,
			"Error":           c.Query("error"),
		})
	})
	r.POST("/admin/register", registerLimiter, func(c *gin.Context) {
		msg := "self-service admin registration is disabled"
		if wantsHTMLResponse(c) {
			c.Redirect(http.StatusFound, "/admin/register?error="+msg)
			return
		}
		c.JSON(http.StatusNotImplemented, gin.H{"ok": false, "error": msg})
	})

	admin := r.Group("/admin", auth.AdminRequired())
	{
		admin.GET("", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin_dashboard.tmpl", gin.H{
				"SiteName":        cfg.SiteName,
				"SiteDescription": cfg.SiteDescription,
				"Env":             cfg.Env,
				"BaseURL":         cfg.BaseURL,
				"User":            cfg.AdminUser,
			})
		})

		admin.POST("/posts", func(c *gin.Context) {
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
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, row)
		})

		admin.PUT("/posts/:slug", func(c *gin.Context) {
			slug := c.Param("slug")
			var body struct {
				Title     string `json:"title" binding:"required"`
				Summary   string `json:"summary"`
				ContentMD string `json:"content_md" binding:"required"`
				CoverURL  string `json:"cover_url"`
				Status    string `json:"status"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, row)
		})

		admin.DELETE("/posts/:slug", func(c *gin.Context) {
			slug := c.Param("slug")
			ctx := c.Request.Context()
			if err := postSvc.Delete(ctx, slug); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		admin.POST("/posts/:slug/categories/:cat", func(c *gin.Context) {
			ctx := c.Request.Context()
			if err := postSvc.AddCategory(ctx, c.Param("slug"), c.Param("cat")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})
		admin.DELETE("/posts/:slug/categories/:cat", func(c *gin.Context) {
			ctx := c.Request.Context()
			if err := postSvc.RemoveCategory(ctx, c.Param("slug"), c.Param("cat")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		admin.POST("/posts/:slug/tags/:tag", func(c *gin.Context) {
			ctx := c.Request.Context()
			if err := postSvc.AddTag(ctx, c.Param("slug"), c.Param("tag")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})
		admin.DELETE("/posts/:slug/tags/:tag", func(c *gin.Context) {
			ctx := c.Request.Context()
			if err := postSvc.RemoveTag(ctx, c.Param("slug"), c.Param("tag")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		admin.POST("/categories", func(c *gin.Context) {
			var body struct{ Name, Slug string }
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx := c.Request.Context()
			row, err := queries.CreateCategory(ctx, appdb.CreateCategoryParams{Name: body.Name, Slug: body.Slug})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, row)
		})
		admin.DELETE("/categories/:slug", func(c *gin.Context) {
			ctx := c.Request.Context()
			if err := queries.DeleteCategoryBySlug(ctx, c.Param("slug")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		admin.POST("/tags", func(c *gin.Context) {
			var body struct{ Name, Slug string }
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx := c.Request.Context()
			row, err := queries.CreateTag(ctx, appdb.CreateTagParams{Name: body.Name, Slug: body.Slug})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, row)
		})
		admin.DELETE("/tags/:slug", func(c *gin.Context) {
			ctx := c.Request.Context()
			if err := queries.DeleteTagBySlug(ctx, c.Param("slug")); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
		})
	}
}

func verifyArgon2idHash(encoded, password string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid argon2 hash format")
	}
	if parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported argon2 variant: %s", parts[1])
	}
	if !strings.HasPrefix(parts[2], "v=") {
		return false, fmt.Errorf("invalid argon2 version segment: %s", parts[2])
	}
	memory, iterations, parallelism, err := parseArgon2Params(parts[3])
	if err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}
	computed := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expected)))
	if len(computed) != len(expected) {
		return false, fmt.Errorf("argon2 hash length mismatch")
	}
	return subtle.ConstantTimeCompare(computed, expected) == 1, nil
}

func parseArgon2Params(segment string) (memory uint32, iterations uint32, parallelism uint8, err error) {
	fields := strings.Split(segment, ",")
	for _, field := range fields {
		kv := strings.SplitN(field, "=", 2)
		if len(kv) != 2 {
			return 0, 0, 0, fmt.Errorf("invalid argon2 param: %s", field)
		}
		value, parseErr := strconv.ParseUint(kv[1], 10, 32)
		if parseErr != nil {
			return 0, 0, 0, fmt.Errorf("parse argon2 param %s: %w", kv[0], parseErr)
		}
		switch kv[0] {
		case "m":
			memory = uint32(value)
		case "t":
			iterations = uint32(value)
		case "p":
			parallelism = uint8(value)
		}
	}
	if memory == 0 || iterations == 0 || parallelism == 0 {
		return 0, 0, 0, fmt.Errorf("argon2 parameters incomplete")
	}
	return memory, iterations, parallelism, nil
}

func wantsHTMLResponse(c *gin.Context) bool {
	ct := c.GetHeader("Content-Type")
	if strings.Contains(ct, "application/x-www-form-urlencoded") || strings.Contains(ct, "multipart/form-data") {
		return true
	}
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "text/html")
}
