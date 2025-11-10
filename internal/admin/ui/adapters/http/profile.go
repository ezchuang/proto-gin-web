package adminuihttp

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"proto-gin-web/internal/domain"
	"proto-gin-web/internal/infrastructure/platform"
	"proto-gin-web/internal/interfaces/http/view"
	"proto-gin-web/internal/interfaces/http/view/presenter"
)

func registerProfileRoutes(group *gin.RouterGroup, cfg platform.Config, adminSvc domain.AdminService) {
	group.GET("/profile", func(c *gin.Context) {
		isForm := view.WantsHTMLResponse(c)
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

		presenter.AdminProfilePage(c, cfg, profile, c.Query("updated") == "1", c.Query("error"))
	})

	group.POST("/profile", func(c *gin.Context) {
		isForm := view.WantsHTMLResponse(c)
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
				presenter.AdminProfileError(c, cfg, current.Email, req.DisplayName, message, status)
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
}
