package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"event-gateway/internal/config"
	"event-gateway/internal/handlers"
	"event-gateway/internal/publisher"
)
import "github.com/prometheus/client_golang/prometheus/promhttp"

// Server wraps the HTTP server used by the application.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	publisher  *publisher.KafkaPublisher
}

// New creates a new HTTP server instance.
func New(cfg config.Config, logger *slog.Logger) *Server {
	pub := publisher.NewKafkaPublisher(cfg.KafkaBrokers)

mux := http.NewServeMux()

mux.HandleFunc("GET /health", handlers.Health)
mux.HandleFunc("POST /webhook", handlers.Webhook(pub))
mux.Handle("GET /metrics", promhttp.Handler())

httpServer := &http.Server{
	Addr:              fmt.Sprintf(":%s", cfg.Port),
	Handler:           mux,
	ReadHeaderTimeout: 5 * time.Second,
}

return &Server{
	httpServer: httpServer,
	logger:     logger,
	publisher:  pub,
}
}

// Start begins accepting HTTP connections.
func (s *Server) Start() error {
	s.logger.Info("starting HTTP server",
		"address", s.httpServer.Addr,
	)

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	if err := s.publisher.Close(); err != nil {
		return err
	}

	return nil
}
