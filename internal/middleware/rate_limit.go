package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/evyataryagoni/ip2country/internal/limiter"
)

// RateLimitMiddleware creates a middleware that enforces rate limiting per IP address
// Returns 429 Too Many Requests when the rate limit is exceeded
//
// Parameters:
//   - limiter: the rate limiter implementation (memory or Redis)
//
// Returns:
//   - func(http.Handler) http.Handler: middleware function
func RateLimitMiddleware(lim limiter.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the real client IP (handles proxies, load balancers, etc.)
			// Chi's RealIP middleware should be applied before this
			ip := r.RemoteAddr

			// Try to get real IP from headers (for proxies/load balancers)
			// Priority: X-Real-IP > X-Forwarded-For > RemoteAddr
			if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				ip = realIP
			} else if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				// X-Forwarded-For can contain multiple IPs, take the first one
				// Format: "client, proxy1, proxy2"
				if firstIP := forwardedFor; firstIP != "" {
					ip = firstIP
				}
			}

			// Check if request is allowed
			if !lim.Allow(ip) {
				// Rate limit exceeded - return 429 Too Many Requests
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Rate limit exceeded. Please try again later.",
				})
				return
			}

			// Request is allowed - continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}
