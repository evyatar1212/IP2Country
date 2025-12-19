package limiter

import (
	"fmt"
	"strings"
)

// LimiterConfig holds configuration for creating a rate limiter
type LimiterConfig struct {
	Type              string  // "memory" or "redis"
	RequestsPerSecond float64 // Rate limit (can be fractional, e.g., 0.2 = 1 req per 5 sec)

	// Redis-specific config
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

// NewLimiter creates a rate limiter based on the configuration (factory pattern)
func NewLimiter(cfg LimiterConfig) (Limiter, error) {
	limiterType := strings.ToLower(strings.TrimSpace(cfg.Type))

	switch limiterType {
	case "memory", "":
		// In-memory rate limiter (good for single-server deployments)
		return NewMemoryLimiter(cfg.RequestsPerSecond), nil

	case "redis":
		// Redis-based rate limiter (required for multi-server deployments)
		limiter, err := NewRedisLimiter(
			cfg.RedisAddr,
			cfg.RedisPassword,
			cfg.RedisDB,
			cfg.RequestsPerSecond,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Redis limiter: %w", err)
		}
		return limiter, nil

	default:
		return nil, fmt.Errorf("unknown rate limiter type: %s (supported: 'memory', 'redis')", cfg.Type)
	}
}
