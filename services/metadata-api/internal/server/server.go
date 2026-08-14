package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"metadata-api/internal/config"
	"metadata-api/internal/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct {
	httpServer *http.Server
	db         *sql.DB
	logger     *slog.Logger
}

func New(cfg config.Config, db *sql.DB, logger *slog.Logger) *Server {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("GET /events", handlers.Events(db))
	mux.Handle("GET /metrics", promhttp.Handler())

	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%s", cfg.Port),
		Handler: otelhttp.NewHandler(
			mux,
			"http-server",
		),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		db: db,
		logger: logger,
	}
}

func (s *Server) Run() error {
	s.logger.Info(
		"starting HTTP server",
		"address",
		s.httpServer.Addr,
	)

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	return s.db.Close()
}