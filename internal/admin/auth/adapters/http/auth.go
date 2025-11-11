package authhttp

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	authdomain "proto-gin-web/internal/admin/auth/domain"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/http/view"
	"proto-gin-web/internal/interfaces/http/view/presenter"
)

func RegisterRoutes(r *gin.Engine, cfg platform.Config, adminSvc authdomain.AdminService, loginLimiter, registerLimiter gin.HandlerFunc) {
	r.GET("/admin/login", func(c *gin.Context) {
		presenter.AdminLoginPage(c, cfg, c.Query("error"))
	})

	r.POST("/admin/login", loginLimiter, func(c *gin.Context) {
		emailInput := c.PostForm("u")
		password := c.PostForm("p")
		isForm := view.WantsHTMLResponse(c)

		ctx := c.Request.Context()
		account, err := adminSvc.Login(ctx, authdomain.AdminLoginInput{
			Email:    emailInput,
			Password: password,
		})
		if err != nil {
			normalized := authdomain.NormalizeEmail(emailInput)
			if errors.Is(err, authdomain.ErrAdminInvalidCredentials) {
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
		if view.WantsHTMLResponse(c) {
			c.Redirect(http.StatusFound, "/admin/login")
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.GET("/admin/register", func(c *gin.Context) {
		presenter.AdminRegisterPage(c, cfg, c.Query("error"))
	})

	r.POST("/admin/register", registerLimiter, func(c *gin.Context) {
		isForm := view.WantsHTMLResponse(c)

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
		created, err := adminSvc.Register(ctx, authdomain.AdminRegisterInput{
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
			case errors.Is(err, authdomain.ErrAdminInvalidEmail):
				status = http.StatusBadRequest
				msg = "invalid email address"
			case errors.Is(err, authdomain.ErrAdminPasswordTooShort):
				status = http.StatusBadRequest
				msg = "password must be at least 8 characters"
			case errors.Is(err, authdomain.ErrAdminPasswordMismatch):
				status = http.StatusBadRequest
				msg = "passwords do not match"
			case errors.Is(err, authdomain.ErrAdminEmailExists):
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
}
