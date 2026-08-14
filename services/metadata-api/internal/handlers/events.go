package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"metadata-api/internal/database"
	"time"

	"metadata-api/internal/metrics"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)



func Events(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		metrics.APIRequestsTotal.Inc()

		defer func() {
			metrics.RequestDuration.Observe(
				time.Since(start).Seconds(),
			)
		}()

		events, err := database.GetEvents(
		r.Context(),
		db,
		)
		if err != nil {

			metrics.APIErrorsTotal.Inc()

			http.Error(
				w,
				"failed to query events",
				http.StatusInternalServerError,
			)

			return
		}

		w.Header().Set("Content-Type", "application/json")

		tracer := otel.Tracer("metadata-api")

		_, span := tracer.Start(
			r.Context(),
			"response.encode_json",
		)
		defer span.End()

		if err := json.NewEncoder(w).Encode(events); err != nil {

			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			metrics.APIErrorsTotal.Inc()

			http.Error(
				w,
				"failed to encode response",
				http.StatusInternalServerError,
			)

			return
		}
	}
}