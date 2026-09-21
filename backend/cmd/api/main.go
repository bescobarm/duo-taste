package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bescobarm/duo-taste/backend/internal/config"
	"github.com/bescobarm/duo-taste/backend/internal/httpapi"
	"github.com/bescobarm/duo-taste/backend/internal/store"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpapi.NewRouter(store.NewMemory(), cfg, logger),
		ReadHeaderTimeout: 10 * time.Second,
	}

	stopped := make(chan struct{})
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		<-signals

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			logger.Error("graceful shutdown failed", "error", err)
		}

		close(stopped)
	}()

	logger.Info("duotaste api listening", "addr", cfg.Addr)

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}

	<-stopped
	logger.Info("bye")
}
