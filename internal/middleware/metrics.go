package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/evyataryagoni/ip2country/internal/metrics"
)

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// MetricsMiddleware records HTTP metrics for each request
func MetricsMiddleware(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the response writer to capture status code and size
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status
				size:           0,
			}

			// Record request size
			requestSize := float64(r.ContentLength)
			if requestSize > 0 {
				m.HTTPRequestSize.WithLabelValues(
					r.Method,
					r.URL.Path,
				).Observe(requestSize)
			}

			// Process the request
			next.ServeHTTP(rw, r)

			// Calculate duration
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(rw.statusCode)

			// Record metrics
			m.HTTPRequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				status,
			).Inc()

			m.HTTPRequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
				status,
			).Observe(duration)

			m.HTTPResponseSize.WithLabelValues(
				r.Method,
				r.URL.Path,
				status,
			).Observe(float64(rw.size))
		})
	}
}
