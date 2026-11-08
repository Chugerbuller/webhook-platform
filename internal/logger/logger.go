package logger

import (
	"log/slog"
	"os"

	"github.com/chugerbuller/webhook-platform/internal/config"
)

func NewStdoutLogger(cfg *config.Config) *slog.Logger {
	opts := slog.HandlerOptions{Level: cfg.LogLevel}

	var h slog.Handler = slog.NewJSONHandler(os.Stdout, &opts)
	if cfg.LogFormat == "text" {
		h = slog.NewTextHandler(os.Stdout, &opts)
	}
	logger := slog.New(h)

	return logger
}
