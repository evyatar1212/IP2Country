package service

import (
	"fmt"
	"testing"

	"github.com/evyataryagoni/ip2country/internal/store"
)

// TestIPService_LookupIP_Success tests successful IP lookup
func TestIPService_LookupIP_Success(t *testing.T) {
	tests := []struct {
		name            string
		ip              string
		expectedCity    string
		expectedCountry string
	}{
		{
			name:            "Google DNS",
			ip:              "8.8.8.8",
			expectedCity:    "Mountain View",
			expectedCountry: "United States",
		},
		{
			name:            "Cloudflare DNS",
			ip:              "1.1.1.1",
			expectedCity:    "Sydney",
			expectedCountry: "Australia",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockStore := store.NewMockStore()
			service := NewIPService(mockStore, nil, nil)

			// Act
			result, err := service.LookupIP(tt.ip)

			// Assert
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if result == nil {
				t.Fatal("expected result, got nil")
			}
			if result.City != tt.expectedCity {
				t.Errorf("expected city %s, got %s", tt.expectedCity, result.City)
			}
			if result.Country != tt.expectedCountry {
				t.Errorf("expected country %s, got %s", tt.expectedCountry, result.Country)
			}

			// Verify store was called correctly
			if len(mockStore.FindByIPCalls) != 1 {
				t.Errorf("expected 1 store call, got %d", len(mockStore.FindByIPCalls))
			}
			if mockStore.FindByIPCalls[0] != tt.ip {
				t.Errorf("expected store called with %s, got %s", tt.ip, mockStore.FindByIPCalls[0])
			}
		})
	}
}

// TestIPService_LookupIP_InvalidIP tests validation errors
func TestIPService_LookupIP_InvalidIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{"empty string", ""},
		{"invalid format", "not-an-ip"},
		{"incomplete IPv4", "192.168.1"},
		{"invalid characters", "192.168.1.abc"},
		{"too many octets", "192.168.1.1.1"},
		{"negative numbers", "192.-168.1.1"},
		{"out of range", "300.300.300.300"},
		{"just dots", "..."},
		{"missing octets", "192.168..1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := store.NewMockStore()
			service := NewIPService(mockStore, nil, nil)

			result, err := service.LookupIP(tt.ip)

			if err == nil {
				t.Error("expected validation error, got nil")
			}
			if result != nil {
				t.Error("expected nil result, got data")
			}
			if err.Error() != "invalid IP address format" {
				t.Errorf("expected 'invalid IP address format', got %s", err.Error())
			}

			// Verify store was NOT called for invalid IPs
			if len(mockStore.FindByIPCalls) != 0 {
				t.Errorf("expected 0 store calls for invalid IP, got %d", len(mockStore.FindByIPCalls))
			}
		})
	}
}

// TestIPService_LookupIP_NotFound tests IP not found scenario
func TestIPService_LookupIP_NotFound(t *testing.T) {
	mockStore := store.NewMockStore()
	service := NewIPService(mockStore, nil, nil)

	result, err := service.LookupIP("192.168.1.1")

	if err == nil {
		t.Error("expected not found error, got nil")
	}
	if result != nil {
		t.Error("expected nil result, got data")
	}
	if err.Error() != "IP address not found" {
		t.Errorf("expected 'IP address not found', got %s", err.Error())
	}

	// Verify store was called (validation passed, but not found in store)
	if len(mockStore.FindByIPCalls) != 1 {
		t.Errorf("expected 1 store call, got %d", len(mockStore.FindByIPCalls))
	}
}

// TestIPService_LookupIP_StoreError tests store errors
func TestIPService_LookupIP_StoreError(t *testing.T) {
	mockStore := store.NewMockStore()
	mockStore.FindByIPError = fmt.Errorf("database connection failed")
	service := NewIPService(mockStore, nil, nil)

	result, err := service.LookupIP("8.8.8.8")

	if err == nil {
		t.Error("expected store error, got nil")
	}
	if result != nil {
		t.Error("expected nil result, got data")
	}
	if err.Error() != "database connection failed" {
		t.Errorf("expected 'database connection failed', got %s", err.Error())
	}

	// Verify store was called
	if len(mockStore.FindByIPCalls) != 1 {
		t.Errorf("expected 1 store call, got %d", len(mockStore.FindByIPCalls))
	}
}

// TestIPService_Close tests cleanup
func TestIPService_Close(t *testing.T) {
	mockStore := store.NewMockStore()
	service := NewIPService(mockStore, nil, nil)

	err := service.Close()

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !mockStore.CloseCalled {
		t.Error("expected store Close to be called")
	}
}

// TestIPService_Close_WithError tests close with error
func TestIPService_Close_WithError(t *testing.T) {
	mockStore := store.NewMockStore()
	mockStore.CloseError = fmt.Errorf("failed to close connection")
	service := NewIPService(mockStore, nil, nil)

	err := service.Close()

	if err == nil {
		t.Error("expected error from close, got nil")
	}
	if err.Error() != "failed to close connection" {
		t.Errorf("expected 'failed to close connection', got %s", err.Error())
	}
	if !mockStore.CloseCalled {
		t.Error("expected store Close to be called despite error")
	}
}

// TestIPService_ValidIPv4 tests various valid IPv4 formats
func TestIPService_ValidIPv4(t *testing.T) {
	tests := []string{
		"0.0.0.0",       // Min IP
		"255.255.255.255", // Max IP
		"127.0.0.1",     // Localhost
		"10.0.0.1",      // Private
		"172.16.0.1",    // Private
		"192.168.0.1",   // Private
	}

	for _, ip := range tests {
		t.Run(ip, func(t *testing.T) {
			mockStore := store.NewMockStore()
			service := NewIPService(mockStore, nil, nil)

			// These are valid IPs, they should pass validation
			// (even if not found in store)
			_, err := service.LookupIP(ip)

			// Should not be a validation error
			if err != nil && err.Error() == "invalid IP address format" {
				t.Errorf("valid IPv4 %s rejected by validator", ip)
			}

			// Verify validation passed (store was called)
			if len(mockStore.FindByIPCalls) != 1 {
				t.Errorf("expected store to be called for valid IP")
			}
		})
	}
}

// TestIPService_ValidIPv6 tests IPv6 support
func TestIPService_ValidIPv6(t *testing.T) {
	tests := []string{
		"2001:4860:4860::8888", // Google DNS IPv6
		"::1",                   // Localhost
		"fe80::1",               // Link-local
		"2001:db8::1",           // Documentation
		"::ffff:192.0.2.1",      // IPv4-mapped
	}

	for _, ip := range tests {
		t.Run(ip, func(t *testing.T) {
			mockStore := store.NewMockStore()
			service := NewIPService(mockStore, nil, nil)

			// Should validate successfully (even if not found in store)
			_, err := service.LookupIP(ip)

			// Should not be a validation error
			if err != nil && err.Error() == "invalid IP address format" {
				t.Errorf("valid IPv6 %s rejected by validator", ip)
			}

			// Verify validation passed (store was called)
			if len(mockStore.FindByIPCalls) != 1 {
				t.Errorf("expected store to be called for valid IP")
			}
		})
	}
}

// TestIPService_LookupIP_EmptyStore tests behavior with empty store
func TestIPService_LookupIP_EmptyStore(t *testing.T) {
	mockStore := store.NewEmptyMockStore()
	service := NewIPService(mockStore, nil, nil)

	result, err := service.LookupIP("8.8.8.8")

	if err == nil {
		t.Error("expected not found error, got nil")
	}
	if result != nil {
		t.Error("expected nil result, got data")
	}
	if err.Error() != "IP address not found" {
		t.Errorf("expected 'IP address not found', got %s", err.Error())
	}
}

// TestIPService_MultipleSequentialLookups tests multiple lookups don't interfere
func TestIPService_MultipleSequentialLookups(t *testing.T) {
	mockStore := store.NewMockStore()
	service := NewIPService(mockStore, nil, nil)

	// First lookup
	result1, err1 := service.LookupIP("8.8.8.8")
	if err1 != nil {
		t.Fatalf("first lookup failed: %v", err1)
	}
	if result1.City != "Mountain View" {
		t.Errorf("first lookup: expected Mountain View, got %s", result1.City)
	}

	// Second lookup (different IP)
	result2, err2 := service.LookupIP("1.1.1.1")
	if err2 != nil {
		t.Fatalf("second lookup failed: %v", err2)
	}
	if result2.City != "Sydney" {
		t.Errorf("second lookup: expected Sydney, got %s", result2.City)
	}

	// Third lookup (not found)
	result3, err3 := service.LookupIP("192.168.1.1")
	if err3 == nil {
		t.Error("third lookup: expected not found error")
	}
	if result3 != nil {
		t.Error("third lookup: expected nil result")
	}

	// Verify all lookups were tracked
	if len(mockStore.FindByIPCalls) != 3 {
		t.Errorf("expected 3 store calls, got %d", len(mockStore.FindByIPCalls))
	}
}

// TestIPService_NilMetrics tests service works without metrics
func TestIPService_NilMetrics(t *testing.T) {
	mockStore := store.NewMockStore()
	service := NewIPService(mockStore, nil, nil) // nil metrics

	result, err := service.LookupIP("8.8.8.8")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	// Should work fine without metrics
}
