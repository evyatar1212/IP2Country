package middleware

import (
	"net/http"
	"time"

	"github.com/evyataryagoni/ip2country/internal/logger"
	"github.com/go-chi/chi/v5/middleware"
)

// LoggingMiddleware logs HTTP requests with structured data
func LoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Get request ID from context (set by chi's RequestID middleware)
			requestID := middleware.GetReqID(r.Context())

			// Log request start
			log.Info().
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Msg("Request started")

			// Process request
			next.ServeHTTP(ww, r)

			// Calculate duration
			duration := time.Since(start)

			// Determine log level based on status code
			logEvent := log.Info()
			if ww.Status() >= 500 {
				logEvent = log.Error()
			} else if ww.Status() >= 400 {
				logEvent = log.Warn()
			}

			// Log request completion
			logEvent.
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Dur("duration_ms", duration).
				Msg("Request completed")
		})
	}
}
