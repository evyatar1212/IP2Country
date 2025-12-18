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

// NewLimiter creates a rate limiter based on the provided configuration
// This factory function allows easy switching between in-memory and Redis implementations
//
// Parameters:
//   - cfg: configuration specifying which type of limiter to create
//
// Returns:
//   - Limiter: the created rate limiter (either MemoryLimiter or RedisLimiter)
//   - error: any error that occurred during creation
func NewLimiter(cfg LimiterConfig) (Limiter, error) {
	// Normalize type to lowercase for case-insensitive comparison
	limiterType := strings.ToLower(strings.TrimSpace(cfg.Type))

	switch limiterType {
	case "memory", "":
		// In-memory rate limiter (default)
		// Good for single-server deployments or development
		return NewMemoryLimiter(cfg.RequestsPerSecond), nil

	case "redis":
		// Redis-based rate limiter
		// Required for multi-server deployments to share rate limit state
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
