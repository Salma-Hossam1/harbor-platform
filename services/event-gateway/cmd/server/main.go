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

	"event-gateway/internal/config"
	"event-gateway/internal/server"
)
import "event-gateway/internal/metrics"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	metrics.Register()

	cfg := config.Load()

	srv := server.New(cfg, logger)

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server exited",
				"error", err,
			)
			os.Exit(1)
		}
	}()

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	<-signalCtx.Done()

	logger.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown failed",
			"error", err,
		)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
