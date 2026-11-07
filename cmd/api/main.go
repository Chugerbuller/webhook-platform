package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/chugerbuller/webhook-platform/internal/cfg"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}
func run() error {
	cfg, err := cfg.LoadConfig()
	if err != nil {
		return err
	}

	opts := slog.HandlerOptions{Level: cfg.LogLevel}

	var h slog.Handler = slog.NewJSONHandler(os.Stdout, &opts)
	if cfg.LogFormat == "text" {
		h = slog.NewTextHandler(os.Stdout, &opts)
	}
	logger := slog.New(h)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health)

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           mux,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	errChan := make(chan error, 1)
	go func() {

		err := http.ListenAndServe(":8080", mux)
		if errors.Is(http.ErrServerClosed, err) {
			logger.Info("server closed")
			errChan <- nil
		} else if err != nil {
			logger.Error("failed to start server", "error", err)
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			fatal(logger, "server error", err)
		}
		return err
	case <-ctx.Done():
		logger.Info("shutting down server")
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown server", "error", err)
		_ = srv.Close()
	}
	return <-errChan

}
func fatal(logger *slog.Logger, msg string, err error) {
	logger.Error(msg, "error", err)
	os.Exit(1)
}
func health(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Health string `json:"health"`
	}{
		Health: "OK",
	}
	bytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}
