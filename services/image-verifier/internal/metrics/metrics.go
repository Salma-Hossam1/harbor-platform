package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	VerificationRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "image_verifier_requests_total",
			Help: "Total number of image verification requests.",
		},
	)

	VerificationFailuresTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "image_verifier_failures_total",
			Help: "Total number of failed image signature verifications.",
		},
	)
)

func Register() {
	prometheus.MustRegister(
		VerificationRequestsTotal,
		VerificationFailuresTotal,
	)
}