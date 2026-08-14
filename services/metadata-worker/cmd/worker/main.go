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

	"metadata-worker/internal/config"
	"metadata-worker/internal/metrics"
	"metadata-worker/internal/worker"
)

func main() {
	metrics.Register()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.Load()

	w, err := worker.New(cfg, logger)
if err != nil {
	logger.Error("failed to initialize worker", "error", err)
	os.Exit(1)
}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	metricsServer := metrics.NewServer(cfg.MetricsPort)

go func() {
	errCh <- w.Run(ctx)
}()

go func() {
	if err := metricsServer.Run(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {

		logger.Error(
			"metrics server stopped",
			"error",
			err,
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

	select {
	case <-signalCtx.Done():
		logger.Info("shutdown signal received")
		cancel()

	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("worker stopped", "error", err)
		}
	}

	if err := w.Shutdown(); err != nil {
		logger.Error("shutdown failed", "error", err)
		os.Exit(1)
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(
	context.Background(),
	5*time.Second,
)
defer cancelShutdown()
	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
	logger.Error(
		"failed to shut down metrics server",
		"error",
		err,
	)
}
	logger.Info("worker stopped")
}