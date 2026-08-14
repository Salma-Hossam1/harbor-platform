package consumer

import (
	"context"
	"log/slog"

	"image-signer/internal/signer"

	"github.com/segmentio/kafka-go"
)

// Consumer reads messages from a Kafka topic.
type Consumer struct {
	reader *kafka.Reader
	logger *slog.Logger
	signer *signer.Signer
}

// New creates a Kafka consumer.
func New(
	brokers string,
	topic string,
	groupID string,
	imageSigner *signer.Signer,
	logger *slog.Logger,
) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader: reader,
		logger: logger,
		signer: imageSigner,
	}
}

// Run continuously consumes Kafka messages.
func (c *Consumer) Run(ctx context.Context) error {
	for {
		message, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			return err
		}

		c.logger.Info(
			"message received",
			"topic", message.Topic,
			"partition", message.Partition,
			"offset", message.Offset,
		)

		if err := c.signer.HandleEvent(ctx, message.Value); err != nil {
			c.logger.Error(
				"failed to process event",
				"error", err,
				"offset", message.Offset,
			)

			continue
		}

		c.logger.Info(
			"event processed successfully",
			"offset", message.Offset,
		)
	}
}

// Close closes the Kafka reader.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
