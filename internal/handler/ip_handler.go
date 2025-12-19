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

// NewIPHandler creates a new IP handler with the given service
func NewIPHandler(service *service.IPService) *IPHandler {
	return &IPHandler{
		service: service,
	}
}

// FindCountry handles GET /v1/find-country?ip=<ip>
// @Summary      Find country by IP address
// @Description  Look up geographic location (city and country) for a given IP address
// @Tags         IP Lookup
// @Accept       json
// @Produce      json
// @Param        ip   query      string  true  "IP address (IPv4 or IPv6)"  example(8.8.8.8)
// @Success      200  {object}   models.IPLocation
// @Failure      400  {object}   models.ErrorResponse  "Invalid IP format"
// @Failure      404  {object}   models.ErrorResponse  "IP not found"
// @Failure      429  {object}   models.ErrorResponse  "Rate limit exceeded"
// @Failure      500  {object}   models.ErrorResponse  "Internal server error"
// @Router       /v1/find-country [get]
func (h *IPHandler) FindCountry(w http.ResponseWriter, r *http.Request) {
	// Step 1: Parse query parameter
	ip := r.URL.Query().Get("ip")

	if ip == "" {
		h.respondError(w, http.StatusBadRequest, "Missing 'ip' query parameter")
		return
	}

	// Step 2: Call service layer
	// The service handles validation and data access
	location, err := h.service.LookupIP(ip)
	if err != nil {
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

// respondJSON writes a JSON response with the given status code
func (h *IPHandler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, we can't change the status code since headers are already sent
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// respondError writes an error response with consistent formatting
func (h *IPHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	h.respondJSON(w, statusCode, models.ErrorResponse{Error: message})
}
