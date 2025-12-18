package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evyataryagoni/ip2country/internal/limiter"
)

// TestRateLimitMiddleware_Allowed tests request allowed
func TestRateLimitMiddleware_Allowed(t *testing.T) {
	mockLimiter := limiter.NewMockLimiter(true) // Allow all

	middleware := RateLimitMiddleware(mockLimiter)

	// Create a test handler that tracks if it was called
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with middleware
	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !nextCalled {
		t.Error("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "success" {
		t.Errorf("expected body 'success', got '%s'", rec.Body.String())
	}
}

// TestRateLimitMiddleware_RateLimited tests request blocked
func TestRateLimitMiddleware_RateLimited(t *testing.T) {
	mockLimiter := limiter.NewMockLimiter(false) // Block all

	middleware := RateLimitMiddleware(mockLimiter)

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if nextCalled {
		t.Error("expected next handler NOT to be called")
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", rec.Code)
	}

	var errResp map[string]string
	json.NewDecoder(rec.Body).Decode(&errResp)

	if errResp["error"] != "Rate limit exceeded. Please try again later." {
		t.Errorf("unexpected error message: %s", errResp["error"])
	}
}

// TestRateLimitMiddleware_IPExtraction tests IP extraction logic
func TestRateLimitMiddleware_IPExtraction(t *testing.T) {
	tests := []struct {
		name          string
		remoteAddr    string
		xRealIP       string
		xForwardedFor string
		expectedIP    string
	}{
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1:12345",
		},
		{
			name:       "X-Real-IP takes priority",
			remoteAddr: "192.168.1.1:12345",
			xRealIP:    "10.0.0.1",
			expectedIP: "10.0.0.1",
		},
		{
			name:          "X-Forwarded-For when no X-Real-IP",
			remoteAddr:    "192.168.1.1:12345",
			xForwardedFor: "10.0.0.2",
			expectedIP:    "10.0.0.2",
		},
		{
			name:          "X-Real-IP over X-Forwarded-For",
			remoteAddr:    "192.168.1.1:12345",
			xRealIP:       "10.0.0.1",
			xForwardedFor: "10.0.0.2",
			expectedIP:    "10.0.0.1",
		},
		{
			name:          "X-Forwarded-For with multiple IPs",
			remoteAddr:    "192.168.1.1:12345",
			xForwardedFor: "10.0.0.3, 10.0.0.4, 10.0.0.5",
			expectedIP:    "10.0.0.3, 10.0.0.4, 10.0.0.5",
		},
		{
			name:       "IPv6 RemoteAddr",
			remoteAddr: "[2001:db8::1]:8080",
			expectedIP: "[2001:db8::1]:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLimiter := limiter.NewMockLimiter(true)
			middleware := RateLimitMiddleware(mockLimiter)

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			// Check that limiter was called with expected IP
			if len(mockLimiter.AllowCalls) != 1 {
				t.Fatalf("expected 1 limiter call, got %d", len(mockLimiter.AllowCalls))
			}
			if mockLimiter.AllowCalls[0] != tt.expectedIP {
				t.Errorf("expected IP %s, limiter called with %s", tt.expectedIP, mockLimiter.AllowCalls[0])
			}
		})
	}
}

// TestRateLimitMiddleware_ContentType tests response headers
func TestRateLimitMiddleware_ContentType(t *testing.T) {
	mockLimiter := limiter.NewMockLimiter(false)
	middleware := RateLimitMiddleware(mockLimiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}
}

// TestRateLimitMiddleware_MultipleRequests tests sequential requests
func TestRateLimitMiddleware_MultipleRequests(t *testing.T) {
	mockLimiter := limiter.NewMockLimiter(true)
	middleware := RateLimitMiddleware(mockLimiter)

	callCount := 0
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(nextHandler)

	// Make 3 requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
	}

	if callCount != 3 {
		t.Errorf("expected next handler called 3 times, got %d", callCount)
	}

	if len(mockLimiter.AllowCalls) != 3 {
		t.Errorf("expected limiter called 3 times, got %d", len(mockLimiter.AllowCalls))
	}
}

// TestRateLimitMiddleware_DifferentIPs tests requests from different IPs
func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	mockLimiter := limiter.NewMockLimiter(true)
	middleware := RateLimitMiddleware(mockLimiter)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(nextHandler)

	ips := []string{"192.168.1.1:12345", "192.168.1.2:12345", "192.168.1.3:12345"}

	for _, ip := range ips {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
	}

	// Verify limiter was called for each IP
	if len(mockLimiter.AllowCalls) != 3 {
		t.Errorf("expected limiter called 3 times, got %d", len(mockLimiter.AllowCalls))
	}

	for i, expectedIP := range ips {
		if mockLimiter.AllowCalls[i] != expectedIP {
			t.Errorf("call %d: expected IP %s, got %s", i, expectedIP, mockLimiter.AllowCalls[i])
		}
	}
}

// TestRateLimitMiddleware_MixedAllowDeny tests both allowed and denied requests
func TestRateLimitMiddleware_MixedAllowDeny(t *testing.T) {
	// Create a limiter that tracks calls and returns alternating results
	mockLimiter := &limiter.MockLimiter{
		AllowCalls: []string{},
	}

	middleware := RateLimitMiddleware(mockLimiter)

	allowedCount := 0
	blockedCount := 0

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedCount++
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(nextHandler)

	// First request - allow
	mockLimiter.AllowResult = true
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	// Second request - block
	mockLimiter.AllowResult = false
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code == http.StatusOK {
		allowedCount++
	} else if rec2.Code == http.StatusTooManyRequests {
		blockedCount++
	}

	// Third request - allow
	mockLimiter.AllowResult = true
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	if allowedCount != 2 {
		t.Errorf("expected 2 allowed requests, got %d", allowedCount)
	}
	if blockedCount != 1 {
		t.Errorf("expected 1 blocked request, got %d", blockedCount)
	}
}

// TestRateLimitMiddleware_EmptyHeaders tests behavior with empty headers
func TestRateLimitMiddleware_EmptyHeaders(t *testing.T) {
	mockLimiter := limiter.NewMockLimiter(true)
	middleware := RateLimitMiddleware(mockLimiter)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Real-IP", "")
	req.Header.Set("X-Forwarded-For", "")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should fall back to RemoteAddr when headers are empty
	if len(mockLimiter.AllowCalls) != 1 {
		t.Fatalf("expected 1 limiter call, got %d", len(mockLimiter.AllowCalls))
	}
	if mockLimiter.AllowCalls[0] != "192.168.1.1:12345" {
		t.Errorf("expected RemoteAddr when headers empty, got %s", mockLimiter.AllowCalls[0])
	}
}

// TestRateLimitMiddleware_JSONResponseFormat tests error response format
func TestRateLimitMiddleware_JSONResponseFormat(t *testing.T) {
	mockLimiter := limiter.NewMockLimiter(false)
	middleware := RateLimitMiddleware(mockLimiter)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if _, exists := response["error"]; !exists {
		t.Error("expected 'error' field in JSON response")
	}

	expectedMsg := "Rate limit exceeded. Please try again later."
	if response["error"] != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, response["error"])
	}
}

// TestRateLimitMiddleware_PreservesNextHandlerResponse tests that allowed requests preserve response
func TestRateLimitMiddleware_PreservesNextHandlerResponse(t *testing.T) {
	mockLimiter := limiter.NewMockLimiter(true)
	middleware := RateLimitMiddleware(mockLimiter)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "test-value")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("custom response"))
	})

	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify next handler's response is preserved
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
	if rec.Header().Get("X-Custom-Header") != "test-value" {
		t.Errorf("expected custom header to be preserved")
	}
	if rec.Body.String() != "custom response" {
		t.Errorf("expected custom response body to be preserved")
	}
}
