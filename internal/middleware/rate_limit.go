package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/evyataryagoni/ip2country/internal/limiter"
)

// RateLimitMiddleware enforces rate limiting per IP address (returns 429 when exceeded)
func RateLimitMiddleware(lim limiter.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			// Try to get real IP from headers (for proxies/load balancers)
			// Priority: X-Real-IP > X-Forwarded-For > RemoteAddr
			if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				ip = realIP
			} else if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				// X-Forwarded-For can contain multiple IPs (format: "client, proxy1, proxy2")
				if firstIP := forwardedFor; firstIP != "" {
					ip = firstIP
				}
			}

			if !lim.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Rate limit exceeded. Please try again later.",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
