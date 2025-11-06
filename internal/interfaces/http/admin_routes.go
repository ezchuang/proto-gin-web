package http

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
	appdb "proto-gin-web/internal/infrastructure/pg"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/auth"
)

// registerAdminRoutes wires admin authentication and CRUD handlers.
func registerAdminRoutes(r *gin.Engine, cfg platform.Config, postSvc domain.PostService, adminSvc domain.AdminService, queries *appdb.Queries) {
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
		emailInput := c.PostForm("u")
		password := c.PostForm("p")
		isForm := wantsHTMLResponse(c)

		ctx := c.Request.Context()
		account, err := adminSvc.Login(ctx, domain.AdminLoginInput{
			Email:    emailInput,
			Password: password,
		})
		if err != nil {
			normalized := domain.NormalizeEmail(emailInput)
			if errors.Is(err, domain.ErrAdminInvalidCredentials) {
				slog.Warn("admin login failed",
					slog.String("user", normalized),
					slog.String("ip", c.ClientIP()))
				if isForm {
					c.Redirect(http.StatusFound, "/admin/login?error=Invalid+credentials")
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"ok": false})
				}
				return
			}
			slog.Error("admin login error",
				slog.String("user", normalized),
				slog.String("ip", c.ClientIP()),
				slog.Any("err", err))
			if isForm {
				c.Redirect(http.StatusFound, "/admin/login?error=Internal+server+error")
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"ok": false})
			}
			return
		}

		accountLabel := account.DisplayName
		if accountLabel == "" {
			accountLabel = account.Email
		}
		secureCookie := cfg.Env == "production"
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("admin", "1", 3600, "/", "", secureCookie, true)
		c.SetCookie("admin_user", accountLabel, 3600, "/", "", secureCookie, true)
		c.SetCookie("admin_email", account.Email, 3600, "/", "", secureCookie, true)
		slog.Info("admin login succeeded",
			slog.String("user", account.Email),
			slog.String("ip", c.ClientIP()))
		if isForm {
			c.Redirect(http.StatusFound, "/admin")
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "user": accountLabel, "email": account.Email})
	})
	r.POST("/admin/logout", func(c *gin.Context) {
		secureCookie := cfg.Env == "production"
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("admin", "", -1, "/", "", secureCookie, true)
		c.SetCookie("admin_user", "", -1, "/", "", secureCookie, true)
		c.SetCookie("admin_email", "", -1, "/", "", secureCookie, true)
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
		req.Email = strings.TrimSpace(req.Email)
		if req.Email == "" || req.Password == "" || req.Confirm == "" {
			msg := "all fields are required"
			if isForm {
				c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(msg))
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": msg})
			return
		}
		ctx := c.Request.Context()
		created, err := adminSvc.Register(ctx, domain.AdminRegisterInput{
			Email:           req.Email,
			Password:        req.Password,
			ConfirmPassword: req.Confirm,
		})
		if err != nil {
			var (
				status = http.StatusInternalServerError
				msg    = "internal server error"
			)
			switch {
			case errors.Is(err, domain.ErrAdminInvalidEmail):
				status = http.StatusBadRequest
				msg = "invalid email address"
			case errors.Is(err, domain.ErrAdminPasswordTooShort):
				status = http.StatusBadRequest
				msg = "password must be at least 8 characters"
			case errors.Is(err, domain.ErrAdminPasswordMismatch):
				status = http.StatusBadRequest
				msg = "passwords do not match"
			case errors.Is(err, domain.ErrAdminEmailExists):
				status = http.StatusConflict
				msg = "account already exists"
				slog.Warn("admin registration insert failed",
					slog.String("user", req.Email),
					slog.String("ip", c.ClientIP()),
					slog.String("reason", msg))
			default:
				slog.Error("admin registration failed",
					slog.String("user", req.Email),
					slog.Any("err", err))
			}
			if isForm {
				c.Redirect(http.StatusFound, "/admin/register?error="+url.QueryEscape(msg))
				return
			}
			c.JSON(status, gin.H{"ok": false, "error": msg})
			return
		}

		slog.Info("admin registration succeeded",
			slog.String("user", created.Email),
			slog.String("ip", c.ClientIP()))

		secureCookie := cfg.Env == "production"
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("admin", "1", 3600, "/", "", secureCookie, true)
		c.SetCookie("admin_user", created.DisplayName, 3600, "/", "", secureCookie, true)
		c.SetCookie("admin_email", created.Email, 3600, "/", "", secureCookie, true)

		if isForm {
			c.Redirect(http.StatusFound, "/admin?registered=1")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ok":     true,
			"user":   gin.H{"email": created.Email, "display_name": created.DisplayName},
			"role":   "admin",
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

		admin.GET("/profile", func(c *gin.Context) {
			isForm := wantsHTMLResponse(c)
			email, err := c.Cookie("admin_email")
			if err != nil || email == "" {
				if isForm {
					c.Redirect(http.StatusFound, "/admin/login?error="+url.QueryEscape("please login first"))
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "unauthorized"})
				}
				return
			}

			ctx := c.Request.Context()
			profile, err := adminSvc.GetProfile(ctx, email)
			if err != nil {
				if errors.Is(err, domain.ErrAdminNotFound) {
					slog.Warn("admin profile not found",
						slog.String("user", email),
						slog.String("ip", c.ClientIP()))
					if isForm {
						c.Redirect(http.StatusFound, "/admin/login?error="+url.QueryEscape("account not found"))
					} else {
						c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "account not found"})
					}
					return
				}
				slog.Error("admin profile lookup failed",
					slog.String("user", email),
					slog.Any("err", err))
				if isForm {
					c.Redirect(http.StatusFound, "/admin/profile?error="+url.QueryEscape("internal server error"))
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "internal server error"})
				}
				return
			}

			renderHTML(c, http.StatusOK, "admin_profile.tmpl", gin.H{
				"Title":           "Account Settings",
				"SiteName":        cfg.SiteName,
				"SiteDescription": cfg.SiteDescription,
				"Env":             cfg.Env,
				"BaseURL":         cfg.BaseURL,
				"Profile":         profile,
				"Updated":         c.Query("updated") == "1",
				"Error":           c.Query("error"),
			})
		})

		admin.POST("/profile", func(c *gin.Context) {
			isForm := wantsHTMLResponse(c)
			email, err := c.Cookie("admin_email")
			if err != nil || email == "" {
				if isForm {
					c.Redirect(http.StatusFound, "/admin/login?error="+url.QueryEscape("please login first"))
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "unauthorized"})
				}
				return
			}

			ctx := c.Request.Context()
			current, err := adminSvc.GetProfile(ctx, email)
			if err != nil {
				if errors.Is(err, domain.ErrAdminNotFound) {
					if isForm {
						c.Redirect(http.StatusFound, "/admin/login?error="+url.QueryEscape("account not found"))
					} else {
						c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "account not found"})
					}
					return
				}
				slog.Error("admin profile lookup failed",
					slog.String("user", email),
					slog.Any("err", err))
				if isForm {
					c.Redirect(http.StatusFound, "/admin/profile?error="+url.QueryEscape("internal server error"))
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "internal server error"})
				}
				return
			}

			type profileRequest struct {
				DisplayName string `json:"display_name" form:"display_name"`
				Password    string `json:"password" form:"password"`
				Confirm     string `json:"confirm" form:"confirm"`
			}
			var req profileRequest
			var bindErr error
			if strings.Contains(c.GetHeader("Content-Type"), "application/json") {
				bindErr = c.ShouldBindJSON(&req)
			} else {
				req.DisplayName = strings.TrimSpace(c.PostForm("display_name"))
				req.Password = c.PostForm("password")
				req.Confirm = c.PostForm("confirm")
			}
			req.DisplayName = strings.TrimSpace(req.DisplayName)

			handleProfileError := func(status int, message string) {
				if isForm {
					renderHTML(c, status, "admin_profile.tmpl", gin.H{
						"Title":           "Account Settings",
						"SiteName":        cfg.SiteName,
						"SiteDescription": cfg.SiteDescription,
						"Env":             cfg.Env,
						"BaseURL":         cfg.BaseURL,
						"Profile": gin.H{
							"Email":       current.Email,
							"DisplayName": req.DisplayName,
						},
						"Error": message,
					})
				} else {
					c.JSON(status, gin.H{"ok": false, "error": message})
				}
			}

			if bindErr != nil {
				handleProfileError(http.StatusBadRequest, bindErr.Error())
				return
			}
			updated, err := adminSvc.UpdateProfile(ctx, current.Email, domain.AdminProfileInput{
				DisplayName:     req.DisplayName,
				Password:        req.Password,
				ConfirmPassword: req.Confirm,
			})
			if err != nil {
				switch {
				case errors.Is(err, domain.ErrAdminDisplayNameRequired):
					handleProfileError(http.StatusBadRequest, "display name is required")
				case errors.Is(err, domain.ErrAdminPasswordMismatch):
					handleProfileError(http.StatusBadRequest, "passwords do not match")
				case errors.Is(err, domain.ErrAdminPasswordTooShort):
					handleProfileError(http.StatusBadRequest, "password must be at least 8 characters")
				default:
					slog.Error("admin profile update failed",
						slog.String("user", current.Email),
						slog.Any("err", err))
					handleProfileError(http.StatusInternalServerError, "failed to update profile")
				}
				return
			}

			secureCookie := cfg.Env == "production"
			c.SetSameSite(http.SameSiteStrictMode)
			c.SetCookie("admin_user", updated.DisplayName, 3600, "/", "", secureCookie, true)
			c.SetCookie("admin_email", updated.Email, 3600, "/", "", secureCookie, true)

			slog.Info("admin profile updated",
				slog.String("user", updated.Email),
				slog.String("ip", c.ClientIP()))

			if isForm {
				c.Redirect(http.StatusFound, "/admin/profile?updated=1")
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"ok": true,
				"user": gin.H{
					"email":        updated.Email,
					"display_name": updated.DisplayName,
				},
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

func wantsHTMLResponse(c *gin.Context) bool {
	ct := c.GetHeader("Content-Type")
	if strings.Contains(ct, "application/x-www-form-urlencoded") || strings.Contains(ct, "multipart/form-data") {
		return true
	}
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "text/html")
}
