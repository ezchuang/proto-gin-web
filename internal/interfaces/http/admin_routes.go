package http

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
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
		renderHTML(c, http.StatusOK, "admin_login.tmpl", gin.H{
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
		accountLabel := ""
		if err == nil {
			if ok, verifyErr := verifyArgon2idHash(userRecord.PasswordHash, p); verifyErr == nil && ok {
				authenticated = true
				accountLabel = userRecord.DisplayName
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
			accountLabel = cfg.AdminUser
		}

		if authenticated {
			if accountLabel == "" {
				accountLabel = u
			}
			secureCookie := cfg.Env == "production"
			c.SetSameSite(http.SameSiteStrictMode)
			c.SetCookie("admin", "1", 3600, "/", "", secureCookie, true)
			c.SetCookie("admin_user", accountLabel, 3600, "/", "", secureCookie, true)
			slog.Info("admin login succeeded",
				slog.String("user", u),
				slog.String("ip", c.ClientIP()))
			if isForm {
				c.Redirect(http.StatusFound, "/admin")
				return
			}
			c.JSON(http.StatusOK, gin.H{"ok": true, "user": accountLabel})
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
		c.SetCookie("admin_user", "", -1, "/", "", secureCookie, true)
		if wantsHTMLResponse(c) {
			c.Redirect(http.StatusFound, "/admin/login")
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.GET("/admin/register", func(c *gin.Context) {
		renderHTML(c, http.StatusOK, "admin_register.tmpl", gin.H{
			"SiteName":        cfg.SiteName,
			"SiteDescription": cfg.SiteDescription,
			"Env":             cfg.Env,
			"BaseURL":         cfg.BaseURL,
			"Error":           c.Query("error"),
		})
	})
	r.POST("/admin/register", registerLimiter, func(c *gin.Context) {
		isForm := wantsHTMLResponse(c)

		type registerRequest struct {
			Email    string `json:"email" form:"u"`
			Password string `json:"password" form:"p"`
			Confirm  string `json:"confirm" form:"confirm"`
		}
		var req registerRequest
		var bindErr error
		if strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			bindErr = c.ShouldBindJSON(&req)
		} else {
			req.Email = strings.TrimSpace(c.PostForm("u"))
			req.Password = c.PostForm("p")
			req.Confirm = c.PostForm("confirm")
		}
		if bindErr != nil {
			if isForm {
				c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(bindErr.Error()))
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": bindErr.Error()})
			return
		}
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		if req.Email == "" || req.Password == "" || req.Confirm == "" {
			msg := "all fields are required"
			if isForm {
				c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(msg))
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": msg})
			return
		}
		if !isLikelyEmail(req.Email) {
			msg := "invalid email address"
			if isForm {
				c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(msg))
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": msg})
			return
		}
		if len(req.Password) < 8 {
			msg := "password must be at least 8 characters"
			if isForm {
				c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(msg))
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": msg})
			return
		}
		if req.Password != req.Confirm {
			msg := "passwords do not match"
			if isForm {
				c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(msg))
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": msg})
			return
		}

		ctx := c.Request.Context()
		if _, err := queries.GetUserByEmail(ctx, req.Email); err == nil {
			msg := "account already exists"
			if isForm {
				c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(msg))
				return
			}
			c.JSON(http.StatusConflict, gin.H{"ok": false, "error": msg})
			return
		} else if !errors.Is(err, pgx.ErrNoRows) {
			slog.Error("admin registration lookup failed",
				slog.String("user", req.Email),
				slog.String("ip", c.ClientIP()),
				slog.Any("err", err))
			msg := "internal server error"
			if isForm {
				c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(msg))
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": msg})
			return
		}

		hash, err := hashArgon2idPassword(req.Password)
		if err != nil {
			slog.Error("argon2 hashing failed",
				slog.String("user", req.Email),
				slog.Any("err", err))
			msg := "internal server error"
			if isForm {
				c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(msg))
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": msg})
			return
		}

		role, err := queries.GetRoleByName(ctx, "admin")
		if err != nil {
			slog.Error("admin role lookup failed",
				slog.String("user", req.Email),
				slog.Any("err", err))
			msg := "internal server error"
			if isForm {
				c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(msg))
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": msg})
			return
		}

		display := deriveDisplayName(req.Email)
		created, err := queries.CreateUser(ctx, appdb.CreateUserParams{
			Email:        req.Email,
			DisplayName:  display,
			PasswordHash: hash,
			RoleID:       &role.ID,
		})
		if err != nil {
			if errors.Is(err, appdb.ErrEmailAlreadyExists) {
				msg := "email already exists"
				slog.Warn("admin registration insert failed",
					slog.String("user", req.Email),
					slog.String("ip", c.ClientIP()),
					slog.String("reason", msg))
				if isForm {
					c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(msg))
					return
				}
				c.JSON(http.StatusConflict, gin.H{"ok": false, "error": msg})
				return
			}
			slog.Error("admin registration insert failed",
				slog.String("user", req.Email),
				slog.Any("err", err))
			msg := "internal server error"
			if isForm {
				c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(msg))
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": msg})
			return
		}

		slog.Info("admin registration succeeded",
			slog.String("user", created.Email),
			slog.String("ip", c.ClientIP()))

		secureCookie := cfg.Env == "production"
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("admin", "1", 3600, "/", "", secureCookie, true)
		c.SetCookie("admin_user", created.DisplayName, 3600, "/", "", secureCookie, true)

		if isForm {
			c.Redirect(http.StatusFound, "/admin?registered=1")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ok":     true,
			"user":   gin.H{"email": created.Email, "display_name": created.DisplayName},
			"role":   role.Name,
			"status": "registered",
		})
	})

	admin := r.Group("/admin", auth.AdminRequired())
	{
		admin.GET("", func(c *gin.Context) {
			userName := cfg.AdminUser
			if v, err := c.Cookie("admin_user"); err == nil && v != "" {
				userName = v
			}
			renderHTML(c, http.StatusOK, "admin_dashboard.tmpl", gin.H{
				"SiteName":        cfg.SiteName,
				"SiteDescription": cfg.SiteDescription,
				"Env":             cfg.Env,
				"BaseURL":         cfg.BaseURL,
				"User":            userName,
				"Registered":      c.Query("registered") == "1",
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

func hashArgon2idPassword(password string) (string, error) {
	const (
		time    = 2
		memory  = 64 * 1024
		threads = 1
		keyLen  = 32
		saltLen = 16
	)
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory, time, threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))
	return encoded, nil
}

func deriveDisplayName(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return email
	}
	local := email[:at]
	parts := strings.FieldsFunc(local, func(r rune) bool {
		return r == '.' || r == '_' || r == '-' || r == '+'
	})
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	if len(parts) == 0 {
		return email
	}
	return strings.Join(parts, " ")
}

func isLikelyEmail(input string) bool {
	if input == "" || strings.Count(input, "@") != 1 {
		return false
	}
	local, domain, ok := strings.Cut(input, "@")
	if !ok || local == "" || domain == "" {
		return false
	}
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") || !strings.Contains(domain, ".") {
		return false
	}
	return true
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
