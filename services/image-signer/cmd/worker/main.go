package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"image-signer/internal/config"
	"image-signer/internal/consumer"
	"image-signer/internal/signer"
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

	logger.Info(
		"starting image signer",
		"kafka_brokers", cfg.KafkaBrokers,
		"kafka_topic", cfg.KafkaTopic,
		"kafka_group_id", cfg.KafkaGroupID,
		"harbor_registry", cfg.HarborRegistry,
	)

	imageSigner := signer.New(
		cfg.CosignBinary,
		cfg.CosignPrivateKey,
		cfg.CosignPassword,
		cfg.HarborRegistry,
		cfg.HarborUsername,
		cfg.HarborPassword,
		cfg.CosignAllowHTTP,
		cfg.CosignIgnoreTLog,
		cfg.CosignAllowInsecure,
	)

	kafkaConsumer := consumer.New(
		cfg.KafkaBrokers,
		cfg.KafkaTopic,
		cfg.KafkaGroupID,
		imageSigner,
		logger,
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	if err := kafkaConsumer.Run(ctx); err != nil {
		logger.Error(
			"consumer stopped unexpectedly",
			"error", err,
		)

		_ = kafkaConsumer.Close()
		os.Exit(1)
	}

	logger.Info("shutdown signal received")

	if err := kafkaConsumer.Close(); err != nil {
		logger.Error(
			"failed to close kafka consumer",
			"error", err,
		)

		os.Exit(1)
	}

	logger.Info("image signer stopped")
}
