package limiter

// MockLimiter is a test double for the Limiter interface
// It allows tests to control allow/deny behavior and verify interactions
type MockLimiter struct {
	// Control behavior
	AllowResult bool // If true, Allow() returns true; if false, returns false

	// Track method calls for verification in tests
	AllowCalls  []string // List of IPs that Allow() was called with
	CloseCalled bool     // Whether Close() was called

	// Control error scenarios
	CloseError error // Error to return from Close(), if any
}

// NewMockLimiter creates a mock limiter with specified allow behavior
// Parameters:
//   - allowResult: if true, all requests will be allowed; if false, all will be denied
func NewMockLimiter(allowResult bool) *MockLimiter {
	return &MockLimiter{
		AllowResult: allowResult,
		AllowCalls:  []string{},
	}
}

// Allow implements the Limiter interface
// Returns the configured AllowResult and tracks the call
func (m *MockLimiter) Allow(ip string) bool {
	m.AllowCalls = append(m.AllowCalls, ip)
	return m.AllowResult
}

// Close implements the Limiter interface
// Tracks that close was called and returns configured error if any
func (m *MockLimiter) Close() error {
	m.CloseCalled = true
	return m.CloseError
}
