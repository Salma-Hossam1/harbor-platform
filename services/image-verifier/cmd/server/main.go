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

	"image-verifier/internal/config"
	"image-verifier/internal/metrics"
	"image-verifier/internal/server"
	"image-verifier/internal/verifier"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	metrics.Register()

	imageVerifier := verifier.New(
		cfg.CosignBinary,
		cfg.CosignPublicKey,
		cfg.HarborRegistry,
		cfg.HarborUsername,
		cfg.HarborPassword,
		cfg.CosignIgnoreTLog,
	)

	httpServer := server.New(
		cfg.Port,
		imageVerifier,
	)

	go func() {
		logger.Info(
			"starting HTTP server",
			"address", httpServer.Addr,
		)

		if err := httpServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Error(
				"HTTP server stopped unexpectedly",
				"error", err,
			)

			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-stop

	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(
			"failed to shut down HTTP server",
			"error", err,
		)

		os.Exit(1)
	}

	logger.Info("server stopped")
}