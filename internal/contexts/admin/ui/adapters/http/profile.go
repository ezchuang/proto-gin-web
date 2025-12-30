package adminuihttp

import (
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

func registerProfileRoutes(group *gin.RouterGroup, cfg config.Config, adminSvc adminusecase.AdminService, sessionMgr *authsession.Manager) {
	group.GET("/profile", func(c *gin.Context) {
		isForm := platformview.WantsHTMLResponse(c)
		profile, ok := adminProfileFromContext(c)
		if !ok {
			if isForm {
				c.Redirect(http.StatusFound, "/admin/login?error="+url.QueryEscape("please login first"))
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "unauthorized"})
			}
			return
		}

		adminview.AdminProfilePage(c, cfg, profile, c.Query("updated") == "1", c.Query("error"))
	})

	group.POST("/profile", func(c *gin.Context) {
		isForm := platformview.WantsHTMLResponse(c)
		sessionProfile, ok := adminProfileFromContext(c)
		if !ok {
			if isForm {
				c.Redirect(http.StatusFound, "/admin/login?error="+url.QueryEscape("please login first"))
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "unauthorized"})
			}
			return
		}

		ctx := c.Request.Context()
		current, err := adminSvc.GetProfile(ctx, sessionProfile.Email)
		if err != nil {
			if errors.Is(err, authdomain.ErrAdminNotFound) {
				if isForm {
					c.Redirect(http.StatusFound, "/admin/login?error="+url.QueryEscape("account not found"))
				} else {
					c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "account not found"})
				}
				return
			}
			slog.Error("admin profile lookup failed",
				slog.String("user", sessionProfile.Email),
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
				adminview.AdminProfileError(c, cfg, current.Email, req.DisplayName, message, status)
			} else {
				c.JSON(status, gin.H{"ok": false, "error": message})
			}
		}

		if bindErr != nil {
			handleProfileError(http.StatusBadRequest, bindErr.Error())
			return
		}

		updated, err := adminSvc.UpdateProfile(ctx, current.Email, authdomain.AdminProfileInput{
			DisplayName:     req.DisplayName,
			Password:        req.Password,
			ConfirmPassword: req.Confirm,
		})
		if err != nil {
			switch {
			case errors.Is(err, authdomain.ErrAdminDisplayNameRequired):
				handleProfileError(http.StatusBadRequest, "display name is required")
			case errors.Is(err, authdomain.ErrAdminPasswordMismatch):
				handleProfileError(http.StatusBadRequest, "passwords do not match")
			case errors.Is(err, authdomain.ErrAdminPasswordTooShort):
				handleProfileError(http.StatusBadRequest, "password must be at least 8 characters")
			default:
				slog.Error("admin profile update failed",
					slog.String("user", current.Email),
					slog.Any("err", err))
				handleProfileError(http.StatusInternalServerError, "failed to update profile")
			}
			return
		}

		refreshAdminCookies(c, cfg, updated)
		if sessionID, cookieErr := c.Cookie(cfg.SessionCookieName); cookieErr == nil && sessionID != "" {
			if err := sessionMgr.RefreshSessionProfile(ctx, sessionID, updated); err != nil {
				slog.Warn("admin session refresh failed",
					slog.String("user", updated.Email),
					slog.Any("err", err))
			}
		}

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
}

func refreshAdminCookies(c *gin.Context, cfg config.Config, admin authdomain.Admin) {
	secureCookie := cfg.Env == "production"
	c.SetSameSite(http.SameSiteStrictMode)
	maxAge := 30 * 60
	// Keep the "friendly" cookies in sync so SSR templates and JS can read the latest admin name/email.
	c.SetCookie("admin_user", admin.DisplayName, maxAge, "/", "", secureCookie, true)
	c.SetCookie("admin_email", admin.Email, maxAge, "/", "", secureCookie, true)
}



