package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ValidationRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "admission_webhook_validation_requests_total",
			Help: "Total number of admission validation requests.",
		},
	)

	ValidationRejectedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "admission_webhook_validation_rejected_total",
			Help: "Total number of rejected admission requests.",
		},
	)

	ValidationErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "admission_webhook_validation_errors_total",
			Help: "Total number of admission validation errors.",
		},
	)
)

func Register() {
	prometheus.MustRegister(
		ValidationRequestsTotal,
		ValidationRejectedTotal,
		ValidationErrorsTotal,
	)
}