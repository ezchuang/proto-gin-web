package public

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	postdomain "proto-gin-web/internal/contexts/blog/post/domain"
	postusecase "proto-gin-web/internal/contexts/blog/post/app"
)

func registerHealthRoutes(r *gin.Engine, postSvc postusecase.PostService) {
	r.GET("/livez", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if _, err := postSvc.ListPublished(ctx, postdomain.ListPostsOptions{Limit: 1}); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}


