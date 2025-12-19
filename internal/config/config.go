package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
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

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	// Load .env file if it exists (for local development)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	return &Config{
		Port: getEnv("PORT", "3000"),

		RateLimitType:   getEnv("RATE_LIMITER_TYPE", "memory"),
		RateLimit:       getEnvAsInt("RATE_LIMIT", 1),
		RateLimitWindow: getEnvAsInt("RATE_LIMIT_WINDOW", 1),

		DatastoreType: getEnv("DATASTORE_TYPE", "csv"),
		DatastorePath: getEnv("DATASTORE_PATH", "./data/ip2country.csv"),

		MySQLDSN: getEnv("MYSQL_DSN", ""),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),
	}
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt reads an environment variable as an integer (returns default if not set or invalid)
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsFloat reads an environment variable as a float64 (returns default if not set or invalid)
func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}

	return value
}
