package worker

import (
	"context"
	"database/sql"

	"log/slog"

	"metadata-worker/internal/config"
	"metadata-worker/internal/consumer"
	"metadata-worker/internal/database"
)

type Worker struct {
	consumer *consumer.Consumer
	db       *sql.DB
}

func New(cfg config.Config, logger *slog.Logger) (*Worker, error) {
	db, err := database.Open(cfg)
	if err != nil {
		return nil, err
	}

	return &Worker{
		consumer: consumer.New(
			cfg.KafkaBrokers,
			cfg.KafkaTopic,
			db,
			logger,
		),
		db: db,
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	return w.consumer.Run(ctx)
}

func (w *Worker) Shutdown() error {
	if err := w.consumer.Close(); err != nil {
		return err
	}

	return w.db.Close()
}