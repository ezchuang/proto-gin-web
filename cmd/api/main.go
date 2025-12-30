package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	authsession "proto-gin-web/internal/admin/auth/session"
	admincontentusecase "proto-gin-web/internal/admin/content/app"
	adminuiusecase "proto-gin-web/internal/admin/ui/app"
	adminusecase "proto-gin-web/internal/admin/auth/app"
	postusecase "proto-gin-web/internal/blog/post/app"
	taxonomyusecase "proto-gin-web/internal/application/taxonomy"
	appdb "proto-gin-web/internal/infrastructure/pg"
	platformlog "proto-gin-web/internal/infrastructure/platform"
	redisstore "proto-gin-web/internal/infrastructure/redis"
	"proto-gin-web/internal/platform/config"
	httpapp "proto-gin-web/internal/platform/http"
)

// @title           Proto Gin Web API
// @version         1.0
// @description     Public endpoints for blog content plus admin content APIs.
// @BasePath        /
// @schemes         http https
func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	log := platformlog.NewLogger(cfg.Env, cfg.LogFile)
	slog.SetDefault(log)

	pool, err := appdb.NewPool(context.Background(), cfg)
	if err != nil {
		log.Error("failed to initialize database pool", slog.Any("err", err))
		os.Exit(1)
	}
	defer pool.Close()

	queries := appdb.New(pool)
	rememberRepo := appdb.NewRememberTokenRepository(pool)
	postRepo := appdb.NewPostRepository(pool)
	postSvc := postusecase.NewService(postRepo)
	adminRepo := appdb.NewAdminAccountRepository(queries)
	adminSvc := adminusecase.NewService(adminRepo, adminusecase.Config{
		AdminRoleName: "admin",
	})
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Error("failed to initialize redis client", slog.Any("err", err))
		os.Exit(1)
	}
	defer redisClient.Close()
	sessionStore := redisstore.NewAdminSessionStore(redisClient)
	sessionManager := authsession.NewManager(sessionStore, rememberRepo, authsession.Config{})
	taxonomyRepo := appdb.NewTaxonomyRepository(queries)
	taxonomySvc := taxonomyusecase.NewService(taxonomyRepo)
	adminContentSvc := admincontentusecase.NewService(postSvc, taxonomySvc)
	adminUISvc := adminuiusecase.NewService(postSvc)

	r := httpapp.NewRouter(cfg, postSvc, adminSvc, adminContentSvc, adminUISvc, sessionManager)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("http server starting", slog.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", slog.Any("err", err))
	}
	log.Info("server stopped")
}


