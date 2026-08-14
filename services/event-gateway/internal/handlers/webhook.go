package handlers

import (
	"io"
	"net/http"

	"event-gateway/internal/publisher"

	"log"
)
import "event-gateway/internal/metrics"

// Webhook returns an HTTP handler that publishes the received payload.
func Webhook(pub publisher.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		defer r.Body.Close()

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		log.Printf("%s\n", payload)

		if err := pub.Publish(payload); err != nil {
			http.Error(w, "failed to process webhook", http.StatusInternalServerError)
			return
		}

		metrics.WebhooksReceivedTotal.Inc()
		w.WriteHeader(http.StatusOK)
	}
}