package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dfodeker/terminus/internal/metrics"
	"github.com/go-chi/chi/v5"
)

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(rec, r)

		route := routePattern(r) // e.g. "/users" or "/stores/{id}"
		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		statusStr := strconv.Itoa(status)

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, route, statusStr).Inc()
		metrics.HTTPDurationSeconds.WithLabelValues(r.Method, route, statusStr).Observe(time.Since(start).Seconds())
		metrics.HTTPResponseSizeBytes.WithLabelValues(r.Method, route, statusStr).Observe(float64(rec.bytes))
	})
}

func routePattern(r *http.Request) string {
	if rc := chi.RouteContext(r.Context()); rc != nil {
		if p := rc.RoutePattern(); p != "" {
			return p
		}
	}
	// fallback (works but may be high-cardinality if paths contain IDs)
	return r.URL.Path
}
