package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evyataryagoni/ip2country/internal/models"
	"github.com/evyataryagoni/ip2country/internal/service"
	"github.com/evyataryagoni/ip2country/internal/store"
)

// TestIPHandler_FindCountry_Success tests successful response
func TestIPHandler_FindCountry_Success(t *testing.T) {
	// Arrange
	mockStore := store.NewMockStore()
	svc := service.NewIPService(mockStore, nil, nil)
	handler := NewIPHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/find-country?ip=8.8.8.8", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.FindCountry(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var location models.IPLocation
	if err := json.NewDecoder(rec.Body).Decode(&location); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if location.City != "Mountain View" {
		t.Errorf("expected city 'Mountain View', got '%s'", location.City)
	}
	if location.Country != "United States" {
		t.Errorf("expected country 'United States', got '%s'", location.Country)
	}
}

// TestIPHandler_FindCountry_MissingParameter tests missing IP parameter
func TestIPHandler_FindCountry_MissingParameter(t *testing.T) {
	mockStore := store.NewMockStore()
	svc := service.NewIPService(mockStore, nil, nil)
	handler := NewIPHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/find-country", nil)
	rec := httptest.NewRecorder()

	handler.FindCountry(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var errResp models.ErrorResponse
	json.NewDecoder(rec.Body).Decode(&errResp)

	if errResp.Error != "Missing 'ip' query parameter" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

// TestIPHandler_FindCountry_EmptyParameter tests empty IP parameter
func TestIPHandler_FindCountry_EmptyParameter(t *testing.T) {
	mockStore := store.NewMockStore()
	svc := service.NewIPService(mockStore, nil, nil)
	handler := NewIPHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/find-country?ip=", nil)
	rec := httptest.NewRecorder()

	handler.FindCountry(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var errResp models.ErrorResponse
	json.NewDecoder(rec.Body).Decode(&errResp)

	if errResp.Error != "Missing 'ip' query parameter" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

// TestIPHandler_FindCountry_InvalidIP tests invalid IP format
func TestIPHandler_FindCountry_InvalidIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{"invalid format", "not-an-ip"},
		{"incomplete", "192.168.1"},
		{"invalid chars", "abc.def.ghi.jkl"},
		{"too many octets", "192.168.1.1.1"},
		{"out of range", "300.300.300.300"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := store.NewMockStore()
			svc := service.NewIPService(mockStore, nil, nil)
			handler := NewIPHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/v1/find-country?ip="+tt.ip, nil)
			rec := httptest.NewRecorder()

			handler.FindCountry(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}

			var errResp models.ErrorResponse
			json.NewDecoder(rec.Body).Decode(&errResp)

			if errResp.Error != "invalid IP address format" {
				t.Errorf("expected validation error, got: %s", errResp.Error)
			}
		})
	}
}

// TestIPHandler_FindCountry_NotFound tests IP not found
func TestIPHandler_FindCountry_NotFound(t *testing.T) {
	mockStore := store.NewMockStore()
	svc := service.NewIPService(mockStore, nil, nil)
	handler := NewIPHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/find-country?ip=192.168.1.1", nil)
	rec := httptest.NewRecorder()

	handler.FindCountry(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}

	var errResp models.ErrorResponse
	json.NewDecoder(rec.Body).Decode(&errResp)

	if errResp.Error != "IP address not found" {
		t.Errorf("expected not found error, got: %s", errResp.Error)
	}
}

// TestIPHandler_FindCountry_InternalError tests store errors
func TestIPHandler_FindCountry_InternalError(t *testing.T) {
	mockStore := store.NewMockStore()
	mockStore.FindByIPError = fmt.Errorf("database connection failed")
	svc := service.NewIPService(mockStore, nil, nil)
	handler := NewIPHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/find-country?ip=8.8.8.8", nil)
	rec := httptest.NewRecorder()

	handler.FindCountry(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	var errResp models.ErrorResponse
	json.NewDecoder(rec.Body).Decode(&errResp)

	// Should return generic error message, not leak internal details
	if errResp.Error != "Internal server error" {
		t.Errorf("expected generic error message, got: %s", errResp.Error)
	}
}

// TestIPHandler_FindCountry_MultipleIPs tests multiple IP lookups
func TestIPHandler_FindCountry_MultipleIPs(t *testing.T) {
	tests := []struct {
		ip              string
		expectedCity    string
		expectedCountry string
		expectedStatus  int
	}{
		{"8.8.8.8", "Mountain View", "United States", http.StatusOK},
		{"1.1.1.1", "Sydney", "Australia", http.StatusOK},
		{"192.168.1.1", "", "", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			mockStore := store.NewMockStore()
			svc := service.NewIPService(mockStore, nil, nil)
			handler := NewIPHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/v1/find-country?ip="+tt.ip, nil)
			rec := httptest.NewRecorder()

			handler.FindCountry(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var location models.IPLocation
				json.NewDecoder(rec.Body).Decode(&location)

				if location.City != tt.expectedCity {
					t.Errorf("expected city '%s', got '%s'", tt.expectedCity, location.City)
				}
				if location.Country != tt.expectedCountry {
					t.Errorf("expected country '%s', got '%s'", tt.expectedCountry, location.Country)
				}
			}
		})
	}
}

// TestIPHandler_FindCountry_ValidIPv6 tests IPv6 support
func TestIPHandler_FindCountry_ValidIPv6(t *testing.T) {
	mockStore := store.NewMockStore()
	svc := service.NewIPService(mockStore, nil, nil)
	handler := NewIPHandler(svc)

	// IPv6 addresses should be validated correctly
	ipv6Addresses := []string{
		"2001:4860:4860::8888",
		"::1",
		"fe80::1",
	}

	for _, ip := range ipv6Addresses {
		t.Run(ip, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/find-country?ip="+ip, nil)
			rec := httptest.NewRecorder()

			handler.FindCountry(rec, req)

			// Should not return validation error (400 with "invalid IP address format")
			if rec.Code == http.StatusBadRequest {
				var errResp models.ErrorResponse
				json.NewDecoder(rec.Body).Decode(&errResp)
				if errResp.Error == "invalid IP address format" {
					t.Errorf("valid IPv6 %s rejected as invalid", ip)
				}
			}

			// Can be 404 (not found in store) but not 400 (invalid format)
			if rec.Code != http.StatusNotFound && rec.Code != http.StatusOK {
				t.Errorf("expected 404 or 200 for valid IPv6, got %d", rec.Code)
			}
		})
	}
}

// TestIPHandler_FindCountry_ContentType tests response headers
func TestIPHandler_FindCountry_ContentType(t *testing.T) {
	tests := []struct {
		name       string
		ip         string
		statusCode int
	}{
		{"success response", "8.8.8.8", http.StatusOK},
		{"error response", "invalid", http.StatusBadRequest},
		{"not found", "192.168.1.1", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := store.NewMockStore()
			svc := service.NewIPService(mockStore, nil, nil)
			handler := NewIPHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/v1/find-country?ip="+tt.ip, nil)
			rec := httptest.NewRecorder()

			handler.FindCountry(rec, req)

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}
		})
	}
}

// TestIPHandler_RespondJSON tests JSON response helper
func TestIPHandler_RespondJSON(t *testing.T) {
	handler := &IPHandler{}
	rec := httptest.NewRecorder()

	// Valid JSON encoding
	handler.respondJSON(rec, http.StatusOK, models.IPLocation{
		City:    "Test City",
		Country: "Test Country",
	})

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var location models.IPLocation
	if err := json.NewDecoder(rec.Body).Decode(&location); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if location.City != "Test City" {
		t.Errorf("expected city 'Test City', got '%s'", location.City)
	}
}

// TestIPHandler_RespondError tests error response helper
func TestIPHandler_RespondError(t *testing.T) {
	handler := &IPHandler{}
	rec := httptest.NewRecorder()

	handler.respondError(rec, http.StatusBadRequest, "Test error message")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var errResp models.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error != "Test error message" {
		t.Errorf("expected 'Test error message', got '%s'", errResp.Error)
	}
}

// TestIPHandler_FindCountry_CaseSensitivity tests if IP lookup is case-sensitive (it shouldn't be)
func TestIPHandler_FindCountry_CaseSensitivity(t *testing.T) {
	// IPv6 can have hex characters (a-f), test case sensitivity
	mockStore := store.NewMockStore()
	// Add an IPv6 address to mock data
	mockStore.Data["2001:db8::1"] = &models.IPLocation{
		IP:      "2001:db8::1",
		City:    "Test City",
		Country: "Test Country",
	}

	svc := service.NewIPService(mockStore, nil, nil)
	handler := NewIPHandler(svc)

	tests := []string{
		"2001:db8::1",   // lowercase
		"2001:DB8::1",   // uppercase
		"2001:Db8::1",   // mixed case
	}

	for _, ip := range tests {
		t.Run(ip, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/find-country?ip="+ip, nil)
			rec := httptest.NewRecorder()

			handler.FindCountry(rec, req)

			// Validation should pass for all variations
			// (may be 404 if store doesn't normalize, but not 400)
			if rec.Code == http.StatusBadRequest {
				var errResp models.ErrorResponse
				json.NewDecoder(rec.Body).Decode(&errResp)
				if errResp.Error == "invalid IP address format" {
					t.Errorf("valid IPv6 variation %s rejected as invalid", ip)
				}
			}
		})
	}
}
