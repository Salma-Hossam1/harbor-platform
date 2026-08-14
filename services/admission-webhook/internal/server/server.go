package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"admission-webhook/internal/handlers"
	"admission-webhook/internal/verifier"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func New(
	port string,
	verifierClient *verifier.Client,
	logger *slog.Logger,
) *http.Server {
	mux := http.NewServeMux()

	mux.Handle(
		"POST /validate",
		handlers.NewValidateHandler(verifierClient, logger),
	)

	mux.HandleFunc(
		"GET /health",
		handlers.Health,
	)

	mux.Handle(
		"GET /metrics",
		promhttp.Handler(),
	)

	return &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}