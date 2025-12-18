package router

import (
	"net/http"

	"github.com/evyataryagoni/ip2country/internal/handler"
	"github.com/evyataryagoni/ip2country/internal/limiter"
	"github.com/evyataryagoni/ip2country/internal/logger"
	custommiddleware "github.com/evyataryagoni/ip2country/internal/middleware"
	"github.com/evyataryagoni/ip2country/internal/metrics"
	v1 "github.com/evyataryagoni/ip2country/internal/router/v1"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "github.com/evyataryagoni/ip2country/docs" // Swagger docs
)

// SetupRouter creates and configures the Chi router with all middleware and routes
// This separates routing logic from the main application setup
//
// Parameters:
//   - ipHandler: the IP lookup handler
//   - rateLimiter: the rate limiter (memory or Redis)
//   - m: metrics collector
//   - log: structured logger
//
// Returns:
//   - chi.Router: configured router ready to use
func SetupRouter(ipHandler *handler.IPHandler, rateLimiter limiter.Limiter, m *metrics.Metrics, log *logger.Logger) chi.Router {
	// Create new Chi router
	r := chi.NewRouter()

	// Apply global middleware - these run on every request
	// Order matters! RequestID should be first, then logging, then rate limiting
	r.Use(middleware.RequestID)                           // Add unique request ID to each request
	r.Use(middleware.RealIP)                              // Get real client IP (handles proxies/load balancers)
	r.Use(custommiddleware.LoggingMiddleware(log))        // Structured logging
	r.Use(middleware.Recoverer)                           // Recover from panics and return 500
	r.Use(custommiddleware.RateLimitMiddleware(rateLimiter)) // Rate limiting per IP
	r.Use(custommiddleware.MetricsMiddleware(m))          // Collect Prometheus metrics

	// Mount v1 API routes under /v1 prefix
	// This allows for API versioning (future: /v2, /v3, etc.)
	r.Mount("/v1", v1.SetupRoutes(ipHandler))

	// Root-level routes (not versioned)
	// Health check endpoint - used by load balancers and monitoring
	r.Get("/health", healthCheckHandler)

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Swagger UI endpoint - API documentation
	// Access at: http://localhost:3000/swagger/index.html
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return r
}

// healthCheckHandler is a simple health check endpoint
// Returns 200 OK if the service is running
// In production, you might want to check database connections, etc.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
