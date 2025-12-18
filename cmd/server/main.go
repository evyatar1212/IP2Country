package main

import (
	"fmt"
	"log"
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

	// Initialize logger
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

	// Initialize the store based on configuration
	var dataStore store.Store
	var err error

	switch cfg.DatastoreType {
	case "csv":
		dataStore, err = store.NewCSVStore(cfg.DatastorePath)
		if err != nil {
			log.Fatalf("Failed to initialize CSV store: %v", err)
		}
		fmt.Println("✅ CSV store initialized")
	case "mysql":
		dataStore, err = store.NewMySQLStore(cfg.MySQLDSN)
		if err != nil {
			log.Fatalf("Failed to initialize MySQL store: %v", err)
		}
		fmt.Println("✅ MySQL store initialized")
	case "redis":
		redisStore, err := store.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		if err != nil {
			log.Fatalf("Failed to initialize Redis store: %v", err)
		}
		fmt.Println("✅ Redis store initialized")

		// Auto-load dummy data if Redis is empty
		isEmpty, err := redisStore.IsEmpty()
		if err != nil {
			log.Printf("⚠️  Warning: Failed to check if Redis is empty: %v", err)
		} else if isEmpty {
			fmt.Println("📦 Redis is empty, loading sample data from CSV...")
			if err := redisStore.LoadFromCSV(cfg.DatastorePath); err != nil {
				log.Printf("⚠️  Warning: Failed to load sample data: %v", err)
			}
		}

		dataStore = redisStore
	default:
		log.Fatalf("Unknown datastore type: %s", cfg.DatastoreType)
	}

	// Ensure store is closed when the program exits
	defer dataStore.Close()

	// Initialize rate limiter based on configuration
	// Calculate effective rate: requests per second
	// Example: 1 request per 5 seconds = 1/5 = 0.2 req/s
	effectiveRate := float64(cfg.RateLimit) / float64(cfg.RateLimitWindow)

	rateLimiter, err := limiter.NewLimiter(limiter.LimiterConfig{
		Type:              cfg.RateLimitType,
		RequestsPerSecond: effectiveRate,
		RedisAddr:         cfg.RedisAddr,
		RedisPassword:     cfg.RedisPassword,
		RedisDB:           cfg.RedisDB,
	})
	if err != nil {
		log.Fatalf("Failed to initialize rate limiter: %v", err)
	}
	defer rateLimiter.Close()
	fmt.Printf("✅ Rate limiter initialized (type: %s, limit: %d req per %d sec = %.2f req/s)\n",
		cfg.RateLimitType, cfg.RateLimit, cfg.RateLimitWindow, effectiveRate)

	// Initialize metrics collector
	m := metrics.New()
	appLogger.Info().Msg("Metrics initialized")

	// Create service layer (business logic)
	ipService := service.NewIPService(dataStore, m, appLogger)
	defer ipService.Close()

	// Create handler layer (HTTP handlers)
	ipHandler := handler.NewIPHandler(ipService)

	// Set up router with all routes and middleware
	r := router.SetupRouter(ipHandler, rateLimiter, m, appLogger)

	// Start the server
	serverAddr := ":" + cfg.Port
	appLogger.Info().
		Str("port", cfg.Port).
		Str("api_endpoint", "http://localhost:"+cfg.Port+"/v1/find-country?ip=<ip>").
		Str("health_check", "http://localhost:"+cfg.Port+"/health").
		Str("metrics", "http://localhost:"+cfg.Port+"/metrics").
		Msg("Server is running")

	// Start HTTP server
	log.Fatal(http.ListenAndServe(serverAddr, r))
}
