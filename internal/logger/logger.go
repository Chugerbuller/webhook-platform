// Package logger предоставляет функции для создания логгера.
package logger

import (
	"log/slog"
	"os"
)
// NewStdoutLogger создает новый логгер, который пишет в stdout.
// Уровень логирования и формат (json или text) задаются параметрами.
func NewStdoutLogger(lvl slog.Level, format string) *slog.Logger {
	opts := slog.HandlerOptions{Level: lvl}

	var h slog.Handler = slog.NewJSONHandler(os.Stdout, &opts)
	if format == "text" {
		h = slog.NewTextHandler(os.Stdout, &opts)
	}
	logger := slog.New(h)

	return logger
}
