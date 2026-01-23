package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

// Request ID context key
type requestIDKey struct{}

// statusRecorder wraps http.ResponseWriter to capture status and bytes written
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Unwrap returns the underlying ResponseWriter (for http.ResponseController compatibility)
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// RequestID middleware generates and stores a unique request ID
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing request ID from upstream proxy
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = generateRequestID()
		}

		// Add to response headers
		w.Header().Set("X-Request-ID", reqID)

		// Store in context
		ctx := context.WithValue(r.Context(), requestIDKey{}, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey{}).(string); ok {
		return reqID
	}
	return ""
}

// generateRequestID creates a random request ID
func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// RequestLogger returns a middleware that logs HTTP request details using slog
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r)

			dur := time.Since(start)
			reqID := GetRequestID(r.Context())
			remoteIP := clientIP(r)
			ua := r.UserAgent()

			// Duration in different formats for flexibility
			durStr := dur.Round(time.Microsecond).String()
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

			// Query string (be careful with sensitive data in production)
			if r.URL.RawQuery != "" {
				attrs = append(attrs, slog.String("query", r.URL.RawQuery))
			}

			// Log at appropriate level based on status code
			if rec.status >= 500 {
				logger.LogAttrs(r.Context(), slog.LevelError, "request completed", attrs...)
			} else if rec.status >= 400 {
				logger.LogAttrs(r.Context(), slog.LevelWarn, "request completed", attrs...)
			} else {
				logger.LogAttrs(r.Context(), slog.LevelInfo, "request completed", attrs...)
			}
		})
	}
}

// clientIP extracts the client IP from the request
func clientIP(r *http.Request) string {
	// Check X-Forwarded-For header (only trust if behind a known proxy)
	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Check X-Real-IP header
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
