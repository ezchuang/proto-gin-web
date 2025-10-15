package platform

import (
	"log/slog"
	"os"
)

func NewLogger(env string) *slog.Logger {
	lvl := slog.LevelInfo
	if env == "development" || env == "dev" || env == "local" {
		lvl = slog.LevelDebug
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}
