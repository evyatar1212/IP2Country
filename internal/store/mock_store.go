package store

import (
	"fmt"

	"github.com/evyataryagoni/ip2country/internal/models"
)

// MockStore is a test double for the Store interface
// It allows tests to control behavior and verify interactions
type MockStore struct {
	// Data holds the mock data (IP address -> location mapping)
	Data map[string]*models.IPLocation

	// Track method calls for verification in tests
	FindByIPCalls []string
	CloseCalled   bool

	// Control behavior for error scenarios
	FindByIPError error
	CloseError    error
}

// NewMockStore creates a mock store with sample test data
// This creates a store pre-populated with common test IPs
func NewMockStore() *MockStore {
	return &MockStore{
		Data: map[string]*models.IPLocation{
			"8.8.8.8": {
				IP:      "8.8.8.8",
				City:    "Mountain View",
				Country: "United States",
			},
			"1.1.1.1": {
				IP:      "1.1.1.1",
				City:    "Sydney",
				Country: "Australia",
			},
		},
		FindByIPCalls: []string{},
	}
}

// NewEmptyMockStore creates a mock store with no data
// Useful for testing "not found" scenarios
func NewEmptyMockStore() *MockStore {
	return &MockStore{
		Data:          map[string]*models.IPLocation{},
		FindByIPCalls: []string{},
	}
}

// FindByIP implements the Store interface
// Tracks calls and returns configured data or errors
func (m *MockStore) FindByIP(ip string) (*models.IPLocation, error) {
	// Track that this method was called with this IP
	m.FindByIPCalls = append(m.FindByIPCalls, ip)

	// If configured to return an error, return it
	if m.FindByIPError != nil {
		return nil, m.FindByIPError
	}

	// Look up the IP in mock data
	location, exists := m.Data[ip]
	if !exists {
		return nil, fmt.Errorf("IP address not found")
	}

	return location, nil
}

// Close implements the Store interface
// Tracks that close was called and returns configured error if any
func (m *MockStore) Close() error {
	m.CloseCalled = true
	return m.CloseError
}
