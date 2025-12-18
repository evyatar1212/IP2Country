package handler

import (
	"encoding/json"
	"net/http"

	"github.com/evyataryagoni/ip2country/internal/models"
	"github.com/evyataryagoni/ip2country/internal/service"
)

// IPHandler handles HTTP requests for IP lookups
// This is the handler layer - it deals with HTTP concerns only
//
// Responsibilities:
//   - Parse HTTP requests (query parameters)
//   - Call service methods
//   - Format HTTP responses (JSON)
//   - Set appropriate status codes
//   - NO business logic (that's in the service layer)
type IPHandler struct {
	service *service.IPService
}

// NewIPHandler creates a new IP handler
// Constructor function that injects the service dependency
//
// Parameters:
//   - service: the IP service that handles business logic
//
// Returns:
//   - *IPHandler: pointer to the created handler
func NewIPHandler(service *service.IPService) *IPHandler {
	return &IPHandler{
		service: service,
	}
}

// FindCountry handles GET /v1/find-country?ip=<ip>
// This is the main API endpoint for IP lookups
//
// Query Parameters:
//   - ip: the IP address to lookup (required)
//
// Responses:
//   - 200: Success with location data
//   - 400: Invalid IP format
//   - 404: IP not found
//   - 500: Internal server error
func (h *IPHandler) FindCountry(w http.ResponseWriter, r *http.Request) {
	// Step 1: Parse query parameter
	// r.URL.Query().Get("ip") extracts the "ip" parameter from the URL
	ip := r.URL.Query().Get("ip")

	// Check if IP parameter is missing
	if ip == "" {
		h.respondError(w, http.StatusBadRequest, "Missing 'ip' query parameter")
		return
	}

	// Step 2: Call service layer
	// The service handles validation and data access
	location, err := h.service.LookupIP(ip)
	if err != nil {
		// Determine the appropriate HTTP status code based on error
		if err.Error() == "invalid IP address format" {
			h.respondError(w, http.StatusBadRequest, err.Error())
		} else if err.Error() == "IP address not found" {
			h.respondError(w, http.StatusNotFound, err.Error())
		} else {
			// Any other error is an internal server error
			h.respondError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Step 3: Return success response
	h.respondJSON(w, http.StatusOK, location)
}

// respondJSON writes a JSON response
// Helper method to avoid repeating JSON encoding logic
//
// Parameters:
//   - w: the response writer
//   - statusCode: HTTP status code (200, 404, etc.)
//   - data: the data to encode as JSON (any type)
func (h *IPHandler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Set status code
	w.WriteHeader(statusCode)

	// Encode data as JSON and write to response
	// json.NewEncoder(w).Encode() handles the conversion to JSON
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, log it (in production, use proper logging)
		// We can't change the status code now since headers are already sent
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// respondError writes an error response
// Helper method for consistent error formatting
//
// Parameters:
//   - w: the response writer
//   - statusCode: HTTP status code (400, 404, 500, etc.)
//   - message: error message to return
func (h *IPHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	h.respondJSON(w, statusCode, models.ErrorResponse{Error: message})
}
