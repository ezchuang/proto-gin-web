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

	postusecase "proto-gin-web/internal/application/post"
	appdb "proto-gin-web/internal/infrastructure/pg"
	"proto-gin-web/internal/infrastructure/platform"
	httpapp "proto-gin-web/internal/interfaces/http"
)

func main() {
	_ = godotenv.Load()

	cfg := platform.Load()

	log := platform.NewLogger(cfg.Env)
	slog.SetDefault(log)

	pool, err := appdb.NewPool(context.Background(), cfg)
	if err != nil {
		log.Error("failed to initialize database pool", slog.Any("err", err))
		os.Exit(1)
	}
	defer pool.Close()

	postRepo := appdb.NewPostRepository(pool)
	postSvc := postusecase.NewService(postRepo)
	queries := appdb.New(pool)

	r := httpapp.NewRouter(cfg, postSvc, queries)

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
