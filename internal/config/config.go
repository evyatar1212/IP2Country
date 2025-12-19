package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
// In Go, we use structs to group related data
type Config struct {
	// Server configuration
	Port string

	// Rate limiting
	RateLimitType   string // "memory" or "redis"
	RateLimit       int    // number of requests allowed
	RateLimitWindow int    // time window in seconds (default: 1)

	// Datastore configuration
	DatastoreType string // "csv", "mysql", or "redis"
	DatastorePath string // path to CSV file

	// MySQL configuration
	MySQLDSN string // Data Source Name

	// Redis configuration
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

// Load reads configuration from environment variables
// with sensible defaults
// This is a function that returns a pointer to Config
func Load() *Config {
	// Load .env file if it exists (for local development)
	// In production/Docker, environment variables are set directly
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	return &Config{
		// Server config with defaults
		Port: getEnv("PORT", "3000"),

		// Rate limiting (default: memory, 10 requests per 1 second)
		RateLimitType:   getEnv("RATE_LIMITER_TYPE", "memory"),
		RateLimit:       getEnvAsInt("RATE_LIMIT", 1),
		RateLimitWindow: getEnvAsInt("RATE_LIMIT_WINDOW", 1), // default 1 second window

		// Datastore config
		DatastoreType: getEnv("DATASTORE_TYPE", "csv"),
		DatastorePath: getEnv("DATASTORE_PATH", "./data/ip2country.csv"),

		// MySQL config
		MySQLDSN: getEnv("MYSQL_DSN", ""),

		// Redis config
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),
	}
}

// getEnv reads an environment variable or returns a default value
// This is a helper function (lowercase = private to this package)
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt reads an environment variable as an integer
// Returns default if not set or invalid
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	// Try to convert string to integer
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		// If conversion fails, return default
		return defaultValue
	}

	return value
}

// getEnvAsFloat reads an environment variable as a float64
// Returns default if not set or invalid
func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	// Try to convert string to float
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		// If conversion fails, return default
		return defaultValue
	}

	return value
}
