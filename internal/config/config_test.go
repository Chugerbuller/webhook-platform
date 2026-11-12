package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

// clearEnv задаёт пустые значения всем переменным конфига,
// чтобы тесты не зависели от окружения машины.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"PORT", "READ_HEADER_TIMEOUT", "READ_TIMEOUT", "WRITE_TIMEOUT",
		"IDLE_TIMEOUT", "SHUTDOWN_TIMEOUT", "LOG_LEVEL", "LOG_FORMAT",
	} {
		t.Setenv(k, "")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "8080")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FORMAT", "json")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Addr() != ":8080" {
		t.Errorf("Addr() = %q, want %q", cfg.Addr(), ":8080")
	}
	checks := []struct {
		name string
		got  time.Duration
		want time.Duration
	}{
		{"ReadHeaderTimeout", cfg.ReadHeaderTimeout, 5 * time.Second},
		{"ReadTimeout", cfg.ReadTimeout, 30 * time.Second},
		{"WriteTimeout", cfg.WriteTimeout, 30 * time.Second},
		{"IdleTimeout", cfg.IdleTimeout, 90 * time.Second},
		{"ShutdownTimeout", cfg.ShutdownTimeout, 15 * time.Second},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelInfo)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("LogFormat = %q, want %q", cfg.LogFormat, "json")
	}
}

func TestLoadConfig_EmptyDurationsUseDefaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "9000")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FORMAT", "json")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("ReadTimeout = %v, want default 30s", cfg.ReadTimeout)
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "9090")
	t.Setenv("READ_HEADER_TIMEOUT", "2s")
	t.Setenv("SHUTDOWN_TIMEOUT", "1m")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "TEXT")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Addr() != ":9090" {
		t.Errorf("Addr() = %q, want %q", cfg.Addr(), ":9090")
	}
	if cfg.ReadHeaderTimeout != 2*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want 2s", cfg.ReadHeaderTimeout)
	}
	if cfg.ShutdownTimeout != time.Minute {
		t.Errorf("ShutdownTimeout = %v, want 1m", cfg.ShutdownTimeout)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("LogFormat = %q, want %q (формат приводится к нижнему регистру)", cfg.LogFormat, "text")
	}
}

func TestLoadConfig_InvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{"порт не число", map[string]string{"PORT": "abc"}, "PORT"},
		{"порт ноль", map[string]string{"PORT": "0"}, "PORT"},
		{"порт больше 65535", map[string]string{"PORT": "70000"}, "PORT"},
		{"кривая длительность", map[string]string{"READ_TIMEOUT": "10 секунд"}, "READ_TIMEOUT"},
		{"нулевой READ_HEADER_TIMEOUT", map[string]string{"READ_HEADER_TIMEOUT": "0s"}, "READ_HEADER_TIMEOUT"},
		{"неизвестный уровень логов", map[string]string{"LOG_LEVEL": "verbose"}, "LOG_LEVEL"},
		{"неизвестный формат логов", map[string]string{"LOG_FORMAT": "xml"}, "LOG_FORMAT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv("PORT", "8080")
			t.Setenv("LOG_LEVEL", "info")
			t.Setenv("LOG_FORMAT", "json")
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := LoadConfig()
			if err == nil {
				t.Fatalf("LoadConfig() = %+v, want error", cfg)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to mention %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig_CollectsAllErrors(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "abc")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FORMAT", "xml")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want error")
	}
	for _, key := range []string{"PORT", "LOG_FORMAT"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error = %q, want it to mention %q", err, key)
		}
	}
}
