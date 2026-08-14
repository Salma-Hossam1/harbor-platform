package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	EventsProcessedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "metadata_events_processed_total",
			Help: "Total number of metadata events successfully stored.",
		},
	)

	EventsFailedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "metadata_events_failed_total",
			Help: "Total number of metadata events that failed processing.",
		},
	)
)

func Register() {
	prometheus.MustRegister(
		EventsProcessedTotal,
		EventsFailedTotal,
	)
}