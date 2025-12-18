package store

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
)

// TestRedisStore_Connection tests Redis connection
func TestRedisStore_Connection(t *testing.T) {
	// Start mock Redis server
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// Connect to mock Redis
	store, err := NewRedisStore(mr.Addr(), "", 0)
	if err != nil {
		t.Fatalf("failed to connect to Redis: %v", err)
	}
	defer store.Close()

	// Verify connection is working
	if store.client == nil {
		t.Error("expected client to be initialized")
	}
}

// TestRedisStore_ConnectionFailure tests connection errors
func TestRedisStore_ConnectionFailure(t *testing.T) {
	_, err := NewRedisStore("invalid:9999", "", 0)

	if err == nil {
		t.Error("expected connection error, got nil")
	}
}

// TestRedisStore_FindByIP_Success tests successful lookup
func TestRedisStore_FindByIP_Success(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStore(mr.Addr(), "", 0)
	defer store.Close()

	// Set test data
	err := store.Set("8.8.8.8", "Mountain View", "United States")
	if err != nil {
		t.Fatalf("failed to set data: %v", err)
	}

	// Lookup
	location, err := store.FindByIP("8.8.8.8")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if location.IP != "8.8.8.8" {
		t.Errorf("expected IP '8.8.8.8', got '%s'", location.IP)
	}
	if location.City != "Mountain View" {
		t.Errorf("expected 'Mountain View', got '%s'", location.City)
	}
	if location.Country != "United States" {
		t.Errorf("expected 'United States', got '%s'", location.Country)
	}
}

// TestRedisStore_FindByIP_NotFound tests IP not found
func TestRedisStore_FindByIP_NotFound(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStore(mr.Addr(), "", 0)
	defer store.Close()

	location, err := store.FindByIP("192.168.1.1")

	if err == nil {
		t.Error("expected not found error, got nil")
	}
	if location != nil {
		t.Error("expected nil location, got data")
	}
	if err.Error() != "IP address not found" {
		t.Errorf("expected 'IP address not found', got '%s'", err.Error())
	}
}

// TestRedisStore_Set tests setting data
func TestRedisStore_Set(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStore(mr.Addr(), "", 0)
	defer store.Close()

	tests := []struct {
		ip      string
		city    string
		country string
	}{
		{"1.1.1.1", "Sydney", "Australia"},
		{"8.8.8.8", "Mountain View", "United States"},
		{"2.2.2.2", "Paris", "France"},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			err := store.Set(tt.ip, tt.city, tt.country)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify data was stored correctly
			location, err := store.FindByIP(tt.ip)
			if err != nil {
				t.Fatalf("failed to retrieve stored data: %v", err)
			}
			if location.City != tt.city {
				t.Errorf("expected city '%s', got '%s'", tt.city, location.City)
			}
			if location.Country != tt.country {
				t.Errorf("expected country '%s', got '%s'", tt.country, location.Country)
			}
		})
	}
}

// TestRedisStore_Set_Update tests updating existing data
func TestRedisStore_Set_Update(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStore(mr.Addr(), "", 0)
	defer store.Close()

	// Set initial data
	store.Set("8.8.8.8", "Mountain View", "United States")

	// Update with new data
	err := store.Set("8.8.8.8", "San Francisco", "USA")
	if err != nil {
		t.Fatalf("failed to update: %v", err)
	}

	// Verify data was updated
	location, _ := store.FindByIP("8.8.8.8")
	if location.City != "San Francisco" {
		t.Errorf("expected city 'San Francisco', got '%s'", location.City)
	}
	if location.Country != "USA" {
		t.Errorf("expected country 'USA', got '%s'", location.Country)
	}
}

// TestRedisStore_IsEmpty tests empty check
func TestRedisStore_IsEmpty(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStore(mr.Addr(), "", 0)
	defer store.Close()

	// Initially empty
	isEmpty, err := store.IsEmpty()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isEmpty {
		t.Error("expected store to be empty")
	}

	// Add data
	store.Set("8.8.8.8", "Test", "Test")

	// Should not be empty
	isEmpty, err = store.IsEmpty()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isEmpty {
		t.Error("expected store to not be empty")
	}
}

// TestRedisStore_IsEmpty_MultipleKeys tests empty check with multiple keys
func TestRedisStore_IsEmpty_MultipleKeys(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStore(mr.Addr(), "", 0)
	defer store.Close()

	// Add multiple keys
	store.Set("1.1.1.1", "Sydney", "Australia")
	store.Set("8.8.8.8", "Mountain View", "United States")
	store.Set("9.9.9.9", "Berkeley", "United States")

	isEmpty, err := store.IsEmpty()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isEmpty {
		t.Error("expected store with 3 keys to not be empty")
	}
}

// TestRedisStore_Close tests cleanup
func TestRedisStore_Close(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStore(mr.Addr(), "", 0)

	err := store.Close()

	if err != nil {
		t.Errorf("expected no error on close, got: %v", err)
	}
}

// TestRedisStore_Close_NilClient tests close with nil client
func TestRedisStore_Close_NilClient(t *testing.T) {
	store := &RedisStore{client: nil}

	err := store.Close()

	if err != nil {
		t.Errorf("expected no error for nil client, got: %v", err)
	}
}

// TestRedisStore_KeyFormat tests Redis key format
func TestRedisStore_KeyFormat(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStore(mr.Addr(), "", 0)
	defer store.Close()

	// Set a value
	store.Set("8.8.8.8", "Test City", "Test Country")

	// Check that the key exists with the correct format
	val, err := mr.Get("ip:8.8.8.8")
	if err != nil {
		t.Fatalf("expected key 'ip:8.8.8.8' to exist, got error: %v", err)
	}
	if val == "" {
		t.Error("expected key to have value")
	}
}

// TestRedisStore_MultipleIPsIndependent tests that different IPs are independent
func TestRedisStore_MultipleIPsIndependent(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStore(mr.Addr(), "", 0)
	defer store.Close()

	// Set multiple different IPs
	store.Set("8.8.8.8", "Mountain View", "United States")
	store.Set("1.1.1.1", "Sydney", "Australia")
	store.Set("9.9.9.9", "Berkeley", "United States")

	// Verify each one independently
	loc1, _ := store.FindByIP("8.8.8.8")
	if loc1.City != "Mountain View" {
		t.Errorf("IP 8.8.8.8: expected 'Mountain View', got '%s'", loc1.City)
	}

	loc2, _ := store.FindByIP("1.1.1.1")
	if loc2.City != "Sydney" {
		t.Errorf("IP 1.1.1.1: expected 'Sydney', got '%s'", loc2.City)
	}

	loc3, _ := store.FindByIP("9.9.9.9")
	if loc3.City != "Berkeley" {
		t.Errorf("IP 9.9.9.9: expected 'Berkeley', got '%s'", loc3.City)
	}
}

// TestRedisStore_SpecialCharacters tests IPs and values with special characters
func TestRedisStore_SpecialCharacters(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStore(mr.Addr(), "", 0)
	defer store.Close()

	// Test UTF-8 city names
	tests := []struct {
		ip      string
		city    string
		country string
	}{
		{"8.8.8.8", "São Paulo", "Brazil"},
		{"1.1.1.1", "Zürich", "Switzerland"},
		{"2.2.2.2", "北京", "China"},
		{"3.3.3.3", "Москва", "Russia"},
	}

	for _, tt := range tests {
		t.Run(tt.city, func(t *testing.T) {
			// Set with special characters
			err := store.Set(tt.ip, tt.city, tt.country)
			if err != nil {
				t.Fatalf("failed to set data with special chars: %v", err)
			}

			// Retrieve and verify
			location, err := store.FindByIP(tt.ip)
			if err != nil {
				t.Fatalf("failed to retrieve data with special chars: %v", err)
			}
			if location.City != tt.city {
				t.Errorf("expected city '%s', got '%s'", tt.city, location.City)
			}
		})
	}
}

// TestRedisStore_IPv6 tests IPv6 addresses
func TestRedisStore_IPv6(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewRedisStore(mr.Addr(), "", 0)
	defer store.Close()

	ipv6Addresses := []string{
		"2001:4860:4860::8888",
		"::1",
		"fe80::1",
	}

	for _, ip := range ipv6Addresses {
		t.Run(ip, func(t *testing.T) {
			err := store.Set(ip, "Test City", "Test Country")
			if err != nil {
				t.Fatalf("failed to set IPv6: %v", err)
			}

			location, err := store.FindByIP(ip)
			if err != nil {
				t.Fatalf("failed to retrieve IPv6: %v", err)
			}
			if location.IP != ip {
				t.Errorf("expected IP '%s', got '%s'", ip, location.IP)
			}
		})
	}
}
