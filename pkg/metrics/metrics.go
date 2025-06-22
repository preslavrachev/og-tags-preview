package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by endpoint",
		},
		[]string{"endpoint"},
	)

	requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Request latency in seconds by endpoint",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	cacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total cache hits",
	})
	cacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total cache misses",
	})

	responseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_responses_total",
			Help: "HTTP responses by status code and endpoint",
		},
		[]string{"code", "endpoint"},
	)

	circuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "State of the circuit breaker: 0 = closed, 1 = open",
		},
		[]string{"endpoint"},
	)
)

func init() {
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestLatency)
	prometheus.MustRegister(cacheHits, cacheMisses)
	prometheus.MustRegister(responseCounter)
	prometheus.MustRegister(circuitBreakerState)
}

func Inc(endpoint string) {
	requestCounter.WithLabelValues(endpoint).Inc()
}

func Latency(args []string, duration time.Duration) {
	durationSeconds := float64(duration)
	requestLatency.WithLabelValues(args...).Observe(durationSeconds)
}

func CountResponse(code int, endpoint string) {
	responseCounter.WithLabelValues(fmt.Sprintf("%d", code), endpoint).Inc()
}

func SetCBState(endpoint string, open bool) {
	val := 0.0
	if open {
		val = 1.0
	}
	circuitBreakerState.WithLabelValues(endpoint).Set(val)
}

func CacheHit() { cacheHits.Inc() }

func CacheMiss() { cacheMisses.Inc() }
