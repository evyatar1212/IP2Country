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
	appConfig := config.Load()

	// Initialize components
	appLogger := setupLogger(appConfig)
	dataStore := setupDataStore(appConfig, appLogger)
	defer dataStore.Close()

	rateLimiter := setupRateLimiter(appConfig, appLogger)
	defer rateLimiter.Close()

	metricsCollector := setupMetrics(appLogger)

	// Build application layers
	ipService := service.NewIPService(dataStore, metricsCollector, appLogger)
	defer ipService.Close()

	ipHandler := handler.NewIPHandler(ipService)
	appRouter := router.SetupRouter(ipHandler, rateLimiter, metricsCollector, appLogger)

	// Start server
	startServer(appConfig, appRouter, appLogger)
}

// setupLogger initializes the structured logger
func setupLogger(appConfig *config.Config) *logger.Logger {
	appLogger := logger.New(logger.Config{
		Level:  "info",
		Pretty: true,
	})

	appLogger.Info().Msg("Starting IP2Country Server...")
	appLogger.Info().
		Str("port", appConfig.Port).
		Str("rate_limiter_type", appConfig.RateLimitType).
		Int("rate_limit", appConfig.RateLimit).
		Int("rate_limit_window", appConfig.RateLimitWindow).
		Str("datastore_type", appConfig.DatastoreType).
		Str("datastore_path", appConfig.DatastorePath).
		Msg("Configuration loaded")

	return appLogger
}

// setupDataStore initializes the data store based on configuration
// Supports CSV, MySQL, and Redis backends
func setupDataStore(appConfig *config.Config, log *logger.Logger) store.Store {
	var dataStore store.Store
	var err error

	switch appConfig.DatastoreType {
	case "csv":
		dataStore, err = store.NewCSVStore(appConfig.DatastorePath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize CSV store")
		}
		fmt.Println("✅ CSV store initialized")

	case "mysql":
		dataStore, err = store.NewMySQLStore(appConfig.MySQLDSN)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize MySQL store")
		}
		fmt.Println("✅ MySQL store initialized")

	case "redis":
		redisStore, err := store.NewRedisStore(appConfig.RedisAddr, appConfig.RedisPassword, appConfig.RedisDB)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Redis store")
		}
		fmt.Println("✅ Redis store initialized")

		// Auto-load data if Redis is empty
		loadRedisDataIfEmpty(redisStore, appConfig.DatastorePath, log)

		dataStore = redisStore

	default:
		log.Fatal().Str("type", appConfig.DatastoreType).Msg("Unknown datastore type")
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
func setupRateLimiter(appConfig *config.Config, log *logger.Logger) limiter.Limiter {
	// Calculate effective rate: requests per second
	// Example: 10 requests per 5 seconds = 10/5 = 2.0 req/s
	effectiveRate := float64(appConfig.RateLimit) / float64(appConfig.RateLimitWindow)

	rateLimiter, err := limiter.NewLimiter(limiter.LimiterConfig{
		Type:              appConfig.RateLimitType,
		RequestsPerSecond: effectiveRate,
		RedisAddr:         appConfig.RedisAddr,
		RedisPassword:     appConfig.RedisPassword,
		RedisDB:           appConfig.RedisDB,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize rate limiter")
	}

	fmt.Printf("✅ Rate limiter initialized (type: %s, limit: %d req per %d sec = %.2f req/s)\n",
		appConfig.RateLimitType, appConfig.RateLimit, appConfig.RateLimitWindow, effectiveRate)

	return rateLimiter
}

// setupMetrics initializes the Prometheus metrics collector
func setupMetrics(log *logger.Logger) *metrics.Metrics {
	metricsCollector := metrics.New()
	log.Info().Msg("Metrics initialized")
	return metricsCollector
}

// startServer starts the HTTP server and blocks
func startServer(appConfig *config.Config, appRouter http.Handler, log *logger.Logger) {
	serverAddr := ":" + appConfig.Port

	log.Info().
		Str("port", appConfig.Port).
		Str("api_endpoint", "http://localhost:"+appConfig.Port+"/v1/find-country?ip=<ip>").
		Str("health_check", "http://localhost:"+appConfig.Port+"/health").
		Str("metrics", "http://localhost:"+appConfig.Port+"/metrics").
		Str("swagger", "http://localhost:"+appConfig.Port+"/swagger/index.html").
		Msg("Server is running")

	log.Fatal().Err(http.ListenAndServe(serverAddr, appRouter)).Msg("Server failed")
}
