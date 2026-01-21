package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RateLimitExceededTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "courier",
		Name:      "rate_limit_exceeded_total",
		Help:      "Total number of requests rejected by rate limiter",
	})

	GatewayRetriesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "courier",
		Name:      "gateway_retries_total",
		Help:      "Total number of retries when calling external service-order",
	})
)
