package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status"},
	)

	HTTPDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)

	HTTPResponseSizeBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Response size in bytes",
			Buckets: []float64{100, 500, 1_000, 5_000, 10_000, 50_000, 100_000, 500_000, 1_000_000},
		},
		[]string{"method", "route", "status"},
	)
)

func Register(reg prometheus.Registerer) {
	reg.MustRegister(HTTPRequestsTotal, HTTPDurationSeconds, HTTPResponseSizeBytes)
}
