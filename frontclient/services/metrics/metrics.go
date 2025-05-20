package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	// Register metrics
	prometheus.MustRegister(GatewayRequestCounter)
	prometheus.MustRegister(GatewayRequestDuration)
	prometheus.MustRegister(UserRequestCounter)
	prometheus.MustRegister(UserPathCounter)
	prometheus.MustRegister(RateLimitBlockedCounter)

}

var (
	// Request counter
	GatewayRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_request_total",
			Help: "Total number of requests processed by the API handlers, by method and path",
		},
		[]string{"method", "path"},
	)
	// Request duration
	GatewayRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Request duration in seconds for the API handlers, by method and path",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	UserRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_user_request_total",
			Help: "Total requests per user",
		},
		[]string{"user_id"},
	)
	UserPathCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_user_path_total",
			Help: "Total requests paths per user",
		},
		[]string{"user_id", "path"},
	)
	// Метрика для заблокированных запросов
	RateLimitBlockedCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gateway_rate_limit_blocked_total",
			Help: "Total requests blocked by rate limiter",
		},
	)
)
