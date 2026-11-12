package logger

import (
	"context"
	"log/slog"
	"testing"
)

func TestNewStdoutLogger_Handler(t *testing.T) {
	tests := []struct {
		format string
		isJSON bool
	}{
		{"json", true},
		{"text", false},
		{"", true}, // неизвестный формат — по умолчанию JSON
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			l := NewStdoutLogger(slog.LevelInfo, tt.format)
			_, gotJSON := l.Handler().(*slog.JSONHandler)
			if gotJSON != tt.isJSON {
				t.Errorf("format %q: JSON handler = %v, want %v", tt.format, gotJSON, tt.isJSON)
			}
		})
	}
}

func TestNewStdoutLogger_Level(t *testing.T) {
	l := NewStdoutLogger(slog.LevelWarn, "json")
	ctx := context.Background()

	if l.Enabled(ctx, slog.LevelInfo) {
		t.Error("info must be disabled when level is warn")
	}
	if !l.Enabled(ctx, slog.LevelError) {
		t.Error("error must be enabled when level is warn")
	}
}
