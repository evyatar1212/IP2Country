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
func SetupRouter(ipHandler *handler.IPHandler, rateLimiter limiter.Limiter, m *metrics.Metrics, log *logger.Logger) chi.Router {
	r := chi.NewRouter()

	// Apply global middleware (order matters: RequestID → RealIP → Logging → Recoverer → RateLimiting → Metrics)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(custommiddleware.LoggingMiddleware(log))
	r.Use(middleware.Recoverer)
	r.Use(custommiddleware.RateLimitMiddleware(rateLimiter))
	r.Use(custommiddleware.MetricsMiddleware(m))

	// Mount v1 API routes under /v1 prefix (allows future versioning: /v2, /v3, etc.)
	r.Mount("/v1", v1.SetupRoutes(ipHandler))

	// Root-level routes (not versioned)
	r.Get("/health", healthCheckHandler)
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return r
}

// healthCheckHandler returns 200 OK if the service is running
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
