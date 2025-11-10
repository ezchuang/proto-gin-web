package platform

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// NewLogger builds a slog logger that writes to stdout and, if provided, a log file.
func NewLogger(env, logFile string) *slog.Logger {
	lvl := slog.LevelInfo
	if env == "development" || env == "dev" || env == "local" {
		lvl = slog.LevelDebug
	}

	writers := []io.Writer{os.Stdout}
	if logFile != "" {
		if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err == nil {
			if f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
				writers = append(writers, f)
			}
		}
	}

	handler := slog.NewJSONHandler(io.MultiWriter(writers...), &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}
