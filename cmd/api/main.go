package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/chugerbuller/webhook-platform/internal/config"
	"github.com/chugerbuller/webhook-platform/internal/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("api error: %v", err)
	}
}
func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	lg := logger.NewStdoutLogger(cfg.LogLevel, cfg.LogFormat)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           mux,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errChan := make(chan error, 1)
	go func() {
		lg.Info("starting server", "addr", cfg.Addr())
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			errChan <- nil
		} else if err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		lg.Info("shutting down server")
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		errChan <- err
		lg.Info("server shutdown error", "error", err)
		_ = srv.Close()
	}
	return <-errChan

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
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
