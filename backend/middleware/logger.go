package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r)

			dur := time.Since(start)
			reqID := GetRequestID(r.Context())

			// Optional fields (consider privacy + volume):
			remoteIP := clientIP(r)
			ua := r.UserAgent()
			durStr := dur.Round(time.Microsecond).String()

			// numeric ms with decimals (good for grep / dashboards)
			durMs := float64(dur) / float64(time.Millisecond)

			attrs := []slog.Attr{
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int("bytes", rec.bytes),
				slog.String("duration", durStr),
				slog.Float64("duration_ms", durMs),
				slog.String("remote_ip", remoteIP),
			}
			if ua != "" {
				attrs = append(attrs, slog.String("user_agent", ua))
			}

			// Level choice: warn on 5xx, info otherwise (tweak as you like)
			if rec.status >= 500 {
				logger.LogAttrs(r.Context(), slog.LevelWarn, "request completed", attrs...)
			} else {
				logger.LogAttrs(r.Context(), slog.LevelInfo, "request completed", attrs...)
			}
		})
	}
}

func clientIP(r *http.Request) string {
	// If you're behind a trusted proxy, you might use X-Forwarded-For.
	// If you are NOT behind a trusted proxy, do NOT trust these headers.
	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
