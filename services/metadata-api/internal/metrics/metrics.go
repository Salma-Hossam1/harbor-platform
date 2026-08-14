package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	APIRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "metadata_api_requests_total",
			Help: "Total number of Metadata API requests.",
		},
	)

	APIErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "metadata_api_errors_total",
			Help: "Total number of Metadata API 5xx responses.",
		},
	)

	RequestDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "metadata_api_request_duration_seconds",
			Help:    "Request latency for the Metadata API.",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func Register() {
	prometheus.MustRegister(
		APIRequestsTotal,
		APIErrorsTotal,
		RequestDuration,
	)
}