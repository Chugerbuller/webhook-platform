package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port string

	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration

	LogLevel  slog.Level
	LogFormat string // "json" | "text"
}

func (c *Config) Addr() string {
	return ":" + c.Port
}
func LoadConfig() (*Config, error) {

	var (
		cfg  *Config = &Config{}
		errs []string
	)
	fail := func(err error) {
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	cfg.Port = getEnv("PORT", "8080")
	if p, err := strconv.Atoi(cfg.Port); err != nil || p < 1 || p > 65535 {
		fail(fmt.Errorf("PORT: некорректное значение %q", cfg.Port))
	}

	var err error

	cfg.ReadHeaderTimeout, err = getDuration("READ_HEADER_TIMEOUT", 5*time.Second)
	fail(err)
	cfg.ReadTimeout, err = getDuration("READ_TIMEOUT", 30*time.Second)
	fail(err)
	cfg.WriteTimeout, err = getDuration("WRITE_TIMEOUT", 30*time.Second)
	fail(err)
	cfg.IdleTimeout, err = getDuration("IDLE_TIMEOUT", 90*time.Second)
	fail(err)
	cfg.ShutdownTimeout, err = getDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	fail(err)
	if cfg.ReadHeaderTimeout <= 0 {
		fail(fmt.Errorf("READ_HEADER_TIMEOUT: должен быть > 0"))
	}

	if err = cfg.LogLevel.UnmarshalText([]byte(getEnv("LOG_LEVEL", "info"))); err != nil {
		fail(fmt.Errorf("LOG_LEVEL: %w", err))
	}

	cfg.LogFormat = strings.ToLower(getEnv("LOG_FORMAT", "json"))
	if cfg.LogFormat != "json" && cfg.LogFormat != "text" {
		fail(fmt.Errorf("LOG_FORMAT: ожидается json или text, получено %q", cfg.LogFormat))
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("неверная конфигурация: %v", strings.Join(errs, " ; "))
	}

	return cfg, nil
}
func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
func getDuration(key string, def time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return d, nil
}
