package main

import (
	"fmt"
	"net/http"

	"github.com/evyataryagoni/ip2country/internal/config"
	"github.com/evyataryagoni/ip2country/internal/handler"
	"github.com/evyataryagoni/ip2country/internal/limiter"
	"github.com/evyataryagoni/ip2country/internal/logger"
	"github.com/evyataryagoni/ip2country/internal/metrics"
	"github.com/evyataryagoni/ip2country/internal/router"
	"github.com/evyataryagoni/ip2country/internal/service"
	"github.com/evyataryagoni/ip2country/internal/store"
)

// @title           IP2Country API
// @version         1.0
// @description     A high-performance IP geolocation service with rate limiting and multiple storage backends
// @termsOfService  http://swagger.io/terms/

// @contact.name   Evyatar Yagoni
// @contact.email  evyatar@example.com

// @license.name  MIT
// @license.url   http://opensource.org/licenses/MIT

// @host      localhost:3000
// @BasePath  /
func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize components
	appLogger := setupLogger(cfg)
	dataStore := setupDataStore(cfg, appLogger)
	defer dataStore.Close()

	rateLimiter := setupRateLimiter(cfg, appLogger)
	defer rateLimiter.Close()

	metricsCollector := setupMetrics(appLogger)

	// Build application layers
	ipService := service.NewIPService(dataStore, metricsCollector, appLogger)
	defer ipService.Close()

	ipHandler := handler.NewIPHandler(ipService)
	appRouter := router.SetupRouter(ipHandler, rateLimiter, metricsCollector, appLogger)

	// Start server
	startServer(cfg, appRouter, appLogger)
}

// setupLogger initializes the structured logger
func setupLogger(cfg *config.Config) *logger.Logger {
	appLogger := logger.New(logger.Config{
		Level:  "info",
		Pretty: true,
	})

	appLogger.Info().Msg("Starting IP2Country Server...")
	appLogger.Info().
		Str("port", cfg.Port).
		Str("rate_limiter_type", cfg.RateLimitType).
		Int("rate_limit", cfg.RateLimit).
		Int("rate_limit_window", cfg.RateLimitWindow).
		Str("datastore_type", cfg.DatastoreType).
		Str("datastore_path", cfg.DatastorePath).
		Msg("Configuration loaded")

	return appLogger
}

// setupDataStore initializes the data store based on configuration
// Supports CSV, MySQL, and Redis backends
func setupDataStore(cfg *config.Config, log *logger.Logger) store.Store {
	var dataStore store.Store
	var err error

	switch cfg.DatastoreType {
	case "csv":
		dataStore, err = store.NewCSVStore(cfg.DatastorePath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize CSV store")
		}
		fmt.Println("✅ CSV store initialized")

	case "mysql":
		dataStore, err = store.NewMySQLStore(cfg.MySQLDSN)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize MySQL store")
		}
		fmt.Println("✅ MySQL store initialized")

	case "redis":
		redisStore, err := store.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Redis store")
		}
		fmt.Println("✅ Redis store initialized")

		// Auto-load data if Redis is empty
		loadRedisDataIfEmpty(redisStore, cfg.DatastorePath, log)

		dataStore = redisStore

	default:
		log.Fatal().Str("type", cfg.DatastoreType).Msg("Unknown datastore type")
	}

	return dataStore
}

// loadRedisDataIfEmpty checks if Redis is empty and loads sample data from CSV
func loadRedisDataIfEmpty(redisStore *store.RedisStore, csvPath string, log *logger.Logger) {
	isEmpty, err := redisStore.IsEmpty()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to check if Redis is empty")
		return
	}

	if isEmpty {
		fmt.Println("📦 Redis is empty, loading sample data from CSV...")
		if err := redisStore.LoadFromCSV(csvPath); err != nil {
			log.Warn().Err(err).Msg("Failed to load sample data")
		}
	}
}

// setupRateLimiter initializes the rate limiter
// Supports in-memory and Redis-based rate limiting
func setupRateLimiter(cfg *config.Config, log *logger.Logger) limiter.Limiter {
	// Calculate effective rate: requests per second
	// Example: 10 requests per 5 seconds = 10/5 = 2.0 req/s
	effectiveRate := float64(cfg.RateLimit) / float64(cfg.RateLimitWindow)

	rateLimiter, err := limiter.NewLimiter(limiter.LimiterConfig{
		Type:              cfg.RateLimitType,
		RequestsPerSecond: effectiveRate,
		RedisAddr:         cfg.RedisAddr,
		RedisPassword:     cfg.RedisPassword,
		RedisDB:           cfg.RedisDB,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize rate limiter")
	}

	fmt.Printf("✅ Rate limiter initialized (type: %s, limit: %d req per %d sec = %.2f req/s)\n",
		cfg.RateLimitType, cfg.RateLimit, cfg.RateLimitWindow, effectiveRate)

	return rateLimiter
}

// setupMetrics initializes the Prometheus metrics collector
func setupMetrics(log *logger.Logger) *metrics.Metrics {
	metricsCollector := metrics.New()
	log.Info().Msg("Metrics initialized")
	return metricsCollector
}

// startServer starts the HTTP server and blocks
func startServer(cfg *config.Config, appRouter http.Handler, log *logger.Logger) {
	serverAddr := ":" + cfg.Port

	log.Info().
		Str("port", cfg.Port).
		Str("api_endpoint", "http://localhost:"+cfg.Port+"/v1/find-country?ip=<ip>").
		Str("health_check", "http://localhost:"+cfg.Port+"/health").
		Str("metrics", "http://localhost:"+cfg.Port+"/metrics").
		Str("swagger", "http://localhost:"+cfg.Port+"/swagger/index.html").
		Msg("Server is running")

	log.Fatal().Err(http.ListenAndServe(serverAddr, appRouter)).Msg("Server failed")
}
