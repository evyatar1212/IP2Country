package limiter

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLimiter implements distributed rate limiting using Redis
// This is suitable for multi-server deployments where rate limits need to be
// shared across all instances
//
// Algorithm: Token Bucket with Redis
// - Uses Redis keys with TTL for automatic cleanup
// - Uses INCR for atomic counter operations
// - Key format: "ratelimit:{ip}:{window}"
type RedisLimiter struct {
	client         *redis.Client
	ctx            context.Context
	requestsPerSec float64
	windowSize     time.Duration // Time window for rate limiting (e.g., 1 second)
}

// NewRedisLimiter creates a new Redis-based rate limiter
//
// Parameters:
//   - addr: Redis server address (e.g., "localhost:6379")
//   - password: Redis password (empty string if no password)
//   - db: Redis database number (0-15, default is 0)
//   - requestsPerSecond: allowed requests per second per IP (can be fractional, e.g., 0.2)
//
// Returns:
//   - *RedisLimiter: new Redis rate limiter instance
//   - error: any error that occurred during connection
func NewRedisLimiter(addr, password string, db int, requestsPerSecond float64) (*RedisLimiter, error) {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// Test the connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis for rate limiting: %w", err)
	}

	// Calculate appropriate window size based on rate
	// For fractional rates (e.g., 0.2 = 1 req per 5 sec), use longer window
	// For integer rates (e.g., 10 = 10 req per sec), use 1 second window
	windowSize := time.Second
	if requestsPerSecond < 1.0 {
		// For fractional rates, calculate window: 1 / rate
		// Example: 0.2 req/s → 1/0.2 = 5 seconds
		windowSize = time.Duration(float64(time.Second) / requestsPerSecond)
	}

	return &RedisLimiter{
		client:         client,
		ctx:            ctx,
		requestsPerSec: requestsPerSecond,
		windowSize:     windowSize,
	}, nil
}

// Allow checks if a request from the given IP should be allowed
// Uses a Lua script for atomic operations in Redis
//
// How it works:
//  1. Generate a Redis key based on IP and current time window
//  2. Execute a Lua script atomically that:
//     - Increments the counter
//     - Sets expiry if needed
//     - Returns the current count
//  3. Check if count exceeds the limit
//
// Using Lua script ensures atomicity - all operations happen as a single atomic unit
//
// Parameters:
//   - ip: client IP address
//
// Returns:
//   - bool: true if request is allowed, false if rate limited
func (rl *RedisLimiter) Allow(ip string) bool {
	// Generate key based on current time window
	// Format: ratelimit:192.168.1.1:1640000000
	// Window changes based on configured window size (e.g., every 5 seconds for 0.2 req/s)
	now := time.Now()
	windowSeconds := int64(rl.windowSize.Seconds())
	window := now.Unix() / windowSeconds // Rounds down to current window
	key := fmt.Sprintf("ratelimit:%s:%d", ip, window)

	// Lua script for atomic rate limiting
	// This executes atomically on Redis server, no race conditions possible
	luaScript := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local ttl = tonumber(ARGV[2])

		-- Increment the counter atomically
		local current = redis.call('INCR', key)

		-- Set expiry only if this is the first request (count = 1)
		if current == 1 then
			redis.call('EXPIRE', key, ttl)
		end

		-- Return the current count
		return current
	`

	// Execute the Lua script
	// KEYS[1] = key, ARGV[1] = limit, ARGV[2] = TTL in seconds
	result, err := rl.client.Eval(rl.ctx, luaScript, []string{key}, rl.requestsPerSec, int(rl.windowSize.Seconds())*2).Result()
	if err != nil {
		// On Redis error, fail open (allow the request) to avoid blocking legitimate traffic
		// In production, you might want to log this error and use a fallback mechanism
		return true
	}

	// Get the count from Lua script result
	count, ok := result.(int64)
	if !ok {
		// If type assertion fails, fail open
		return true
	}

	// Check if we're within the rate limit
	// For fractional rates, window is adjusted (e.g., 0.2 req/s uses 5-second window)
	// So we allow ceiling of (rate * window) requests per window
	// Example: 0.2 req/s * 5 sec = 1 req per 5-second window
	limit := int64(math.Ceil(rl.requestsPerSec * rl.windowSize.Seconds()))
	return count <= limit
}

// Close closes the Redis connection and cleans up resources
func (rl *RedisLimiter) Close() error {
	if rl.client != nil {
		return rl.client.Close()
	}
	return nil
}
