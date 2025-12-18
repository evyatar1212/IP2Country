package limiter

import (
	"sync"
	"time"
)

// Limiter is the interface that all rate limiters must implement
// This allows us to easily swap between in-memory and Redis implementations
type Limiter interface {
	// Allow checks if a request from the given IP should be allowed
	// Returns true if allowed, false if rate limited
	Allow(ip string) bool

	// Close cleans up any resources (Redis connections, goroutines, etc.)
	Close() error
}

// TokenBucket represents a token bucket for a single client
// The token bucket algorithm allows bursts while maintaining an average rate
//
// How it works:
//   - Each client has a bucket with a maximum capacity
//   - Tokens are added at a fixed rate (e.g., 10/second)
//   - Each request consumes 1 token
//   - If no tokens available, request is rejected (429 Too Many Requests)
type TokenBucket struct {
	tokens         float64   // Current number of tokens in the bucket
	capacity       float64   // Maximum number of tokens (burst size)
	refillRate     float64   // Tokens added per second
	lastRefillTime time.Time // Last time tokens were added
	mu             sync.Mutex // Protects tokens and lastRefillTime
}

// NewTokenBucket creates a new token bucket
//
// Parameters:
//   - rate: tokens per second (e.g., 10 = 10 requests/second)
//   - capacity: maximum tokens (burst size, usually same as rate)
//
// Returns:
//   - *TokenBucket: new token bucket, starts full
func NewTokenBucket(rate float64, capacity float64) *TokenBucket {
	// Start with at least 1 token to allow first request
	// For fractional rates (e.g., 0.2), capacity might be < 1
	initialTokens := capacity
	if initialTokens < 1.0 {
		initialTokens = 1.0
	}

	return &TokenBucket{
		tokens:         initialTokens,
		capacity:       max(capacity, 1.0), // Capacity should be at least 1
		refillRate:     rate,
		lastRefillTime: time.Now(),
	}
}

// Allow checks if a request should be allowed
// This is the main method called for each request
//
// Returns:
//   - bool: true if request is allowed, false if rate limited
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	tb.refill()

	// Check if we have tokens available
	if tb.tokens >= 1.0 {
		// Consume 1 token
		tb.tokens -= 1.0
		return true
	}

	// No tokens available - rate limit exceeded
	return false
}

// refill adds tokens based on time elapsed since last refill
// Must be called with mutex locked
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()

	// Calculate tokens to add: elapsed_time * rate
	// Example: 0.5 seconds * 10 tokens/sec = 5 tokens
	tokensToAdd := elapsed * tb.refillRate

	// Add tokens, but don't exceed capacity
	tb.tokens = min(tb.tokens+tokensToAdd, tb.capacity)

	// Update last refill time
	tb.lastRefillTime = now
}

// MemoryLimiter manages token buckets for multiple clients (per-IP)
// Thread-safe using sync.Map
// This is an in-memory implementation suitable for single-server deployments
type MemoryLimiter struct {
	buckets    sync.Map // map[string]*TokenBucket - keyed by IP address
	rate       float64  // Tokens per second
	capacity   float64  // Maximum tokens (burst size)
	cleanupMu  sync.Mutex
	lastCleanup time.Time
}

// NewMemoryLimiter creates a new in-memory rate limiter
//
// Parameters:
//   - requestsPerSecond: allowed requests per second per IP (can be fractional, e.g., 0.2)
//
// Returns:
//   - *MemoryLimiter: new in-memory rate limiter instance
func NewMemoryLimiter(requestsPerSecond float64) *MemoryLimiter {
	return &MemoryLimiter{
		rate:        requestsPerSecond,
		capacity:    requestsPerSecond, // Burst size equals rate (can burst up to 1 second worth)
		lastCleanup: time.Now(),
	}
}

// Allow checks if a request from the given IP should be allowed
// This is called by the middleware for each request
//
// Parameters:
//   - ip: client IP address
//
// Returns:
//   - bool: true if request is allowed, false if rate limited
func (rl *MemoryLimiter) Allow(ip string) bool {
	// Get or create token bucket for this IP
	bucket := rl.getBucket(ip)

	// Check if request is allowed
	allowed := bucket.Allow()

	// Periodically clean up old buckets (prevent memory leak)
	rl.maybeCleanup()

	return allowed
}

// getBucket gets or creates a token bucket for an IP address
// Thread-safe using sync.Map's LoadOrStore
func (rl *MemoryLimiter) getBucket(ip string) *TokenBucket {
	// Try to load existing bucket
	if value, ok := rl.buckets.Load(ip); ok {
		return value.(*TokenBucket)
	}

	// Create new bucket for this IP
	bucket := NewTokenBucket(rl.rate, rl.capacity)

	// Store it (LoadOrStore handles race conditions)
	actual, _ := rl.buckets.LoadOrStore(ip, bucket)
	return actual.(*TokenBucket)
}

// maybeCleanup periodically removes inactive buckets to prevent memory leak
// Cleans up buckets that haven't been accessed in the last 5 minutes
func (rl *MemoryLimiter) maybeCleanup() {
	rl.cleanupMu.Lock()
	defer rl.cleanupMu.Unlock()

	// Only cleanup every 5 minutes
	if time.Since(rl.lastCleanup) < 5*time.Minute {
		return
	}

	// Cleanup threshold: remove buckets inactive for 5+ minutes
	threshold := time.Now().Add(-5 * time.Minute)

	// Iterate over all buckets
	rl.buckets.Range(func(key, value interface{}) bool {
		bucket := value.(*TokenBucket)
		bucket.mu.Lock()
		lastAccess := bucket.lastRefillTime
		bucket.mu.Unlock()

		// Remove if inactive for too long
		if lastAccess.Before(threshold) {
			rl.buckets.Delete(key)
		}

		return true // continue iteration
	})

	rl.lastCleanup = time.Now()
}

// Close cleans up resources for the in-memory limiter
// For in-memory implementation, there's nothing to clean up
// This method exists to satisfy the Limiter interface
func (rl *MemoryLimiter) Close() error {
	// No resources to clean up for in-memory implementation
	return nil
}

// min returns the minimum of two float64 values
// Helper function for refill logic
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two float64 values
// Helper function for token bucket initialization
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
