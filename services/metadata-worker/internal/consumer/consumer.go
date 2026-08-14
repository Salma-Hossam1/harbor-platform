package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/segmentio/kafka-go"

	"metadata-worker/internal/database"
)
import "metadata-worker/internal/metrics"

type Consumer struct {
	reader *kafka.Reader
	db     *sql.DB
	logger *slog.Logger
}

func New(
	brokers,
	topic string,
	db *sql.DB,
	logger *slog.Logger,
) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  strings.Split(brokers, ","),
			Topic:    topic,
			GroupID:  "metadata-worker",
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		db:     db,
		logger: logger,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
			return err
		}

		var harborEvent database.HarborEvent

		if err := json.Unmarshal(msg.Value, &harborEvent); err != nil {
			metrics.EventsFailedTotal.Inc()
			c.logger.Error(
				"invalid JSON payload",
				"error", err,
			)

			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		if len(harborEvent.EventData.Resources) == 0 {
			metrics.EventsFailedTotal.Inc()
			c.logger.Error(
				"Harbor event contains no resources",
			)

			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		resource := harborEvent.EventData.Resources[0]
		repository := harborEvent.EventData.Repository

		event := database.MetadataEvent{
			EventType:  harborEvent.Type,
			Project:    repository.Namespace,
			Repository: repository.Name,
			Tag:        resource.Tag,
			Digest:     resource.Digest,
			Operator:   harborEvent.Operator,
			OccurredAt: harborEvent.OccurredAt,
		}

		if err := database.InsertEvent(c.db, event); err != nil {
			metrics.EventsFailedTotal.Inc()
			c.logger.Error(
				"failed to insert metadata event",
				"error", err,
			)

			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		metrics.EventsProcessedTotal.Inc()
		c.logger.Info(
			"metadata event inserted",
			"event_type", event.EventType,
			"repository", event.Repository,
		)

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
