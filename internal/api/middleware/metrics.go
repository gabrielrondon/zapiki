package middleware

import (
	"net/http"
	"time"

	"github.com/gabrielrondon/zapiki/internal/metrics"
)

// MetricsMiddleware records metrics for HTTP requests
func MetricsMiddleware(m *metrics.Metrics) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code and size
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call next handler
			next.ServeHTTP(rw, r)

			// Record metrics
			duration := time.Since(start).Seconds()
			requestSize := r.ContentLength
			if requestSize < 0 {
				requestSize = 0
			}

			m.RecordHTTPRequest(
				r.Method,
				r.URL.Path,
				rw.statusCode,
				duration,
				requestSize,
				rw.bytesWritten,
			)
		})
	}
}
