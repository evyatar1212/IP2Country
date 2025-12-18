package limiter

import (
	"sync"
	"testing"
	"time"
)

// TestMemoryLimiter_BasicRateLimit tests basic rate limiting functionality
func TestMemoryLimiter_BasicRateLimit(t *testing.T) {
	// Create a limiter with 5 requests per second
	limiter := NewMemoryLimiter(5)
	defer limiter.Close()

	ip := "192.168.1.1"

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		if !limiter.Allow(ip) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 6th request should be blocked
	if limiter.Allow(ip) {
		t.Error("Request 6 should be rate limited")
	}

	// Wait for refill (1.1 seconds to be safe)
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed again after refill
	if !limiter.Allow(ip) {
		t.Error("Request should be allowed after refill")
	}
}

// TestMemoryLimiter_PerIPIsolation tests that different IPs have separate limits
func TestMemoryLimiter_PerIPIsolation(t *testing.T) {
	limiter := NewMemoryLimiter(3)
	defer limiter.Close()

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Use up limit for IP1
	for i := 0; i < 3; i++ {
		if !limiter.Allow(ip1) {
			t.Errorf("Request %d for IP1 should be allowed", i+1)
		}
	}

	// IP1 should be blocked
	if limiter.Allow(ip1) {
		t.Error("IP1 should be rate limited")
	}

	// IP2 should still be allowed (separate bucket)
	for i := 0; i < 3; i++ {
		if !limiter.Allow(ip2) {
			t.Errorf("Request %d for IP2 should be allowed", i+1)
		}
	}

	// IP2 should now be blocked
	if limiter.Allow(ip2) {
		t.Error("IP2 should be rate limited")
	}
}

// TestMemoryLimiter_Concurrency tests thread safety
func TestMemoryLimiter_Concurrency(t *testing.T) {
	limiter := NewMemoryLimiter(100)
	defer limiter.Close()

	ip := "192.168.1.1"
	allowedCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Spawn 200 goroutines (double the limit)
	// Only 100 should be allowed
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.Allow(ip) {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Should allow around 100 requests (with some tolerance for timing)
	if allowedCount < 95 || allowedCount > 105 {
		t.Errorf("Expected ~100 allowed requests, got %d", allowedCount)
	}
}

// TestMemoryLimiter_TokenRefill tests that tokens refill over time
func TestMemoryLimiter_TokenRefill(t *testing.T) {
	limiter := NewMemoryLimiter(10)
	defer limiter.Close()

	ip := "192.168.1.1"

	// Use up all tokens
	for i := 0; i < 10; i++ {
		limiter.Allow(ip)
	}

	// Should be blocked
	if limiter.Allow(ip) {
		t.Error("Should be rate limited after using all tokens")
	}

	// Wait for partial refill (0.5 seconds = 5 tokens)
	time.Sleep(500 * time.Millisecond)

	// Should allow ~5 more requests
	allowedCount := 0
	for i := 0; i < 10; i++ {
		if limiter.Allow(ip) {
			allowedCount++
		}
	}

	// Should be around 5 (with some tolerance)
	if allowedCount < 4 || allowedCount > 6 {
		t.Errorf("Expected ~5 allowed requests after 0.5s refill, got %d", allowedCount)
	}
}

// TestMemoryLimiter_Close tests that Close doesn't error
func TestMemoryLimiter_Close(t *testing.T) {
	limiter := NewMemoryLimiter(10)

	if err := limiter.Close(); err != nil {
		t.Errorf("Close should not return error, got: %v", err)
	}
}

// TestLimiterInterface_MemoryLimiter tests that MemoryLimiter implements Limiter interface
func TestLimiterInterface_MemoryLimiter(t *testing.T) {
	var _ Limiter = (*MemoryLimiter)(nil)
}

// TestLimiterInterface_RedisLimiter tests that RedisLimiter implements Limiter interface
func TestLimiterInterface_RedisLimiter(t *testing.T) {
	var _ Limiter = (*RedisLimiter)(nil)
}

// TestNewLimiter_Memory tests factory function for memory limiter
func TestNewLimiter_Memory(t *testing.T) {
	tests := []struct {
		name string
		cfg  LimiterConfig
	}{
		{
			name: "explicit memory type",
			cfg: LimiterConfig{
				Type:              "memory",
				RequestsPerSecond: 10,
			},
		},
		{
			name: "uppercase memory type",
			cfg: LimiterConfig{
				Type:              "MEMORY",
				RequestsPerSecond: 10,
			},
		},
		{
			name: "empty type defaults to memory",
			cfg: LimiterConfig{
				Type:              "",
				RequestsPerSecond: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter, err := NewLimiter(tt.cfg)
			if err != nil {
				t.Errorf("NewLimiter() error = %v", err)
				return
			}
			defer limiter.Close()

			// Test that it works
			if !limiter.Allow("192.168.1.1") {
				t.Error("First request should be allowed")
			}
		})
	}
}

// TestNewLimiter_InvalidType tests factory function with invalid type
func TestNewLimiter_InvalidType(t *testing.T) {
	cfg := LimiterConfig{
		Type:              "invalid",
		RequestsPerSecond: 10,
	}

	_, err := NewLimiter(cfg)
	if err == nil {
		t.Error("Expected error for invalid limiter type")
	}
}

// BenchmarkMemoryLimiter_Allow benchmarks the Allow method
func BenchmarkMemoryLimiter_Allow(b *testing.B) {
	limiter := NewMemoryLimiter(1000000) // High limit so we don't hit it
	defer limiter.Close()

	ip := "192.168.1.1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(ip)
	}
}

// BenchmarkMemoryLimiter_AllowParallel benchmarks parallel access
func BenchmarkMemoryLimiter_AllowParallel(b *testing.B) {
	limiter := NewMemoryLimiter(1000000)
	defer limiter.Close()

	b.RunParallel(func(pb *testing.PB) {
		ip := "192.168.1.1"
		for pb.Next() {
			limiter.Allow(ip)
		}
	})
}
