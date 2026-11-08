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
	"time"

	"github.com/chugerbuller/webhook-platform/internal/config"
	"github.com/chugerbuller/webhook-platform/internal/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error in starting server: %v", err)
		return
	}
}
func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	logger := logger.NewStdoutLogger(cfg)

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

		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			logger.Log.Info("server closed")
			errChan <- nil
		} else if err != nil {
			logger.Log.Error("failed to start server", "error", err)
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			logger.Log.Error("server error", "error", err)
		}
		return err
	case <-ctx.Done():
		logger.Log.Info("shutting down server")
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("failed to shutdown server", "error", err)
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
	if _, err := w.Write(bytes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
