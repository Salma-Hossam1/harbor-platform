package server

import (
	"fmt"
	"net/http"
	"time"

	"image-verifier/internal/handlers"
	//"image-verifier/internal/metrics"
	"image-verifier/internal/verifier"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func New(
	port string,
	verifier *verifier.Verifier,
) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /verify", handlers.Verify(verifier))
	mux.Handle("GET /metrics", promhttp.Handler())

	return &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}