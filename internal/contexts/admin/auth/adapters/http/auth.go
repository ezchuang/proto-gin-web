package authhttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	adminusecase "proto-gin-web/internal/contexts/admin/auth/app"
	authdomain "proto-gin-web/internal/contexts/admin/auth/domain"
	authsession "proto-gin-web/internal/contexts/admin/auth/session"
	adminview "proto-gin-web/internal/contexts/admin/ui/adapters/view"
	"proto-gin-web/internal/platform/config"
	platformview "proto-gin-web/internal/platform/http/view"
)

const (
	sessionCookieMaxAge  = 30 * 60
	rememberCookieMaxAge = 30 * 24 * 60 * 60
)

func RegisterRoutes(r *gin.Engine, cfg config.Config, adminSvc adminusecase.AdminService, sessionMgr *authsession.Manager, loginLimiter, registerLimiter gin.HandlerFunc) {
	r.GET("/admin/login", func(c *gin.Context) {
		adminview.AdminLoginPage(c, cfg, c.Query("error"))
	})

	r.POST("/admin/login", loginLimiter, func(c *gin.Context) {
		emailInput := c.PostForm("u")
		password := c.PostForm("p")
		remember := wantsRememberMe(c)
		isForm := platformview.WantsHTMLResponse(c)

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
		session, err := sessionMgr.IssueSession(ctx, account)
		if err != nil {
			slog.Error("admin session issuance failed",
				slog.String("user", account.Email),
				slog.Any("err", err))
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "internal server error"})
			return
		}
		setSessionCookies(c, cfg, session, accountLabel, account.Email)
		if remember {
			if secret, tokenErr := sessionMgr.CreateRememberToken(ctx, account.ID, normalizeDeviceInfo(c.GetHeader("User-Agent"))); tokenErr != nil {
				slog.Error("admin remember token issue failed",
					slog.String("user", account.Email),
					slog.Any("err", tokenErr))
			} else {
				setRememberCookie(c, cfg, secret)
			}
		} else {
			clearRememberState(ctx, c, cfg, sessionMgr)
		}
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
		ctx := c.Request.Context()
		allDevices := c.PostForm("all_devices") == "1" || c.Query("all_devices") == "1"
		sessionID, _ := c.Cookie(cfg.SessionCookieName)
		if sessionID != "" {
			if allDevices {
				if session, err := sessionMgr.ValidateSession(ctx, sessionID); err == nil {
					_ = sessionMgr.DestroyAllSessions(ctx, session.Profile.ID)
					_ = sessionMgr.DeleteRememberTokensByUser(ctx, session.Profile.ID)
				}
			} else {
				_ = sessionMgr.DestroySession(ctx, sessionID)
			}
		}
		if !allDevices {
			clearRememberState(ctx, c, cfg, sessionMgr)
		}
		clearSessionCookies(c, cfg)
		if platformview.WantsHTMLResponse(c) {
			c.Redirect(http.StatusFound, "/admin/login")
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.GET("/admin/register", func(c *gin.Context) {
		adminview.AdminRegisterPage(c, cfg, c.Query("error"))
	})

	r.POST("/admin/register", registerLimiter, func(c *gin.Context) {
		isForm := platformview.WantsHTMLResponse(c)

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

		session, err := sessionMgr.IssueSession(ctx, created)
		if err != nil {
			slog.Error("admin session issuance failed",
				slog.String("user", created.Email),
				slog.Any("err", err))
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "internal server error"})
			return
		}
		setSessionCookies(c, cfg, session, created.DisplayName, created.Email)
		clearRememberState(ctx, c, cfg, sessionMgr)

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

func wantsRememberMe(c *gin.Context) bool {
	v := c.PostForm("remember")
	return v == "1" || strings.EqualFold(v, "on") || strings.EqualFold(v, "true")
}

func setSessionCookies(c *gin.Context, cfg config.Config, session authdomain.AdminSession, displayName, email string) {
	secureCookie := cfg.Env == "production"
	c.SetSameSite(http.SameSiteStrictMode)
	maxAge := sessionCookieMaxAge
	c.SetCookie(cfg.SessionCookieName, session.ID, maxAge, "/", "", secureCookie, true)
	c.SetCookie("admin", "1", maxAge, "/", "", secureCookie, true)
	c.SetCookie("admin_user", displayName, maxAge, "/", "", secureCookie, true)
	c.SetCookie("admin_email", email, maxAge, "/", "", secureCookie, true)
}

func clearSessionCookies(c *gin.Context, cfg config.Config) {
	secureCookie := cfg.Env == "production"
	c.SetSameSite(http.SameSiteStrictMode)
	for _, name := range []string{cfg.SessionCookieName, "admin", "admin_user", "admin_email"} {
		c.SetCookie(name, "", -1, "/", "", secureCookie, true)
	}
	clearRememberCookie(c, cfg)
}

func setRememberCookie(c *gin.Context, cfg config.Config, secret authsession.RememberTokenSecret) {
	secureCookie := cfg.Env == "production"
	c.SetSameSite(http.SameSiteStrictMode)
	value := secret.Selector + ":" + secret.Validator
	maxAge := rememberCookieMaxAge
	c.SetCookie(cfg.RememberCookieName, value, maxAge, "/", "", secureCookie, true)
}

func clearRememberCookie(c *gin.Context, cfg config.Config) {
	secureCookie := cfg.Env == "production"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(cfg.RememberCookieName, "", -1, "/", "", secureCookie, true)
}

func clearRememberState(ctx context.Context, c *gin.Context, cfg config.Config, mgr *authsession.Manager) {
	selector, _ := readRememberSelector(c, cfg)
	if selector != "" {
		_ = mgr.DeleteRememberToken(ctx, selector)
	}
	clearRememberCookie(c, cfg)
}

func readRememberSelector(c *gin.Context, cfg config.Config) (string, string) {
	value, err := c.Cookie(cfg.RememberCookieName)
	if err != nil || value == "" {
		return "", ""
	}
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func normalizeDeviceInfo(userAgent string) string {
	ua := strings.TrimSpace(userAgent)
	if len(ua) > 255 {
		return ua[:255]
	}
	return ua
}



