package metrics

import "github.com/prometheus/client_golang/prometheus"

var WebhooksReceivedTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "harbor_webhooks_received_total",
		Help: "Total number of successfully processed Harbor webhooks.",
	},
)

func Register() {
	prometheus.MustRegister(WebhooksReceivedTotal)
}