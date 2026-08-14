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

	"metadata-api/internal/config"
	"metadata-api/internal/database"
	"metadata-api/internal/server"
	"metadata-api/internal/metrics"

	"metadata-api/internal/telemetry"
)


func main() {

	metrics.Register()
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)

	cfg := config.Load()

	ctx := context.Background()

shutdownTelemetry, err := telemetry.Init(ctx)
if err != nil {
	logger.Error(
		"failed to initialize telemetry",
		"error",
		err,
	)
	os.Exit(1)
}

	db, err := database.Open(cfg)
	if err != nil {
		logger.Error(
			"failed to connect to database",
			"error",
			err,
		)
		os.Exit(1)
	}

	srv := server.New(cfg, db, logger)

	go func() {
		if err := srv.Run(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {

			logger.Error(
				"http server stopped",
				"error",
				err,
			)

			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()

	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error(
			"shutdown failed",
			"error",
			err,
		)
		os.Exit(1)
	}

	if err := shutdownTelemetry(shutdownCtx); err != nil {
	logger.Error(
		"failed to shut down telemetry",
		"error",
		err,
	)
}

	logger.Info("server stopped")
}