package publisher

import (
	"context"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

const HarborEventsTopic = "harbor-events"

// KafkaPublisher publishes webhook payloads to Kafka.
type KafkaPublisher struct {
	writer *kafka.Writer
}

// NewKafkaPublisher creates a Kafka publisher.
func NewKafkaPublisher(brokers string) *KafkaPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(strings.Split(brokers, ",")...),
			Topic:    HarborEventsTopic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

// Publish sends the payload to Kafka.
func (p *KafkaPublisher) Publish(payload []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.writer.WriteMessages(
		ctx,
		kafka.Message{
			Value: payload,
		},
	)
}

// Close releases Kafka resources.
func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}