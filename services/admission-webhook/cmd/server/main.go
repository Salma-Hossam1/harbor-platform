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

	"admission-webhook/internal/config"
	"admission-webhook/internal/metrics"
	"admission-webhook/internal/server"
	"admission-webhook/internal/verifier"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)

	cfg, err := config.Load()
	if err != nil {
		logger.Error(
			"failed to load configuration",
			"error", err,
		)
		os.Exit(1)
	}

	metrics.Register()

	verifierClient := verifier.NewClient(
		cfg.VerifierURL,
	)

	httpServer := server.New(
		cfg.Port,
		verifierClient,
		logger,
	)

	go func() {
		logger.Info(
			"starting HTTPS server",
			"address", httpServer.Addr,
		)

		if err := httpServer.ListenAndServeTLS(
			cfg.TLSCertFile,
			cfg.TLSKeyFile,
		); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {

			logger.Error(
				"HTTPS server stopped unexpectedly",
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