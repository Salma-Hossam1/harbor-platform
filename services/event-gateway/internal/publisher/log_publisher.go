package publisher

import "log/slog"

// LogPublisher logs event payloads.
type LogPublisher struct {
	logger *slog.Logger
}

// NewLogPublisher creates a new LogPublisher.
func NewLogPublisher(logger *slog.Logger) *LogPublisher {
	return &LogPublisher{
		logger: logger,
	}
}

// Publish logs the raw payload.
func (p *LogPublisher) Publish(payload []byte) error {
	p.logger.Info(
		"webhook received",
		"payload", string(payload),
	)

	return nil
}