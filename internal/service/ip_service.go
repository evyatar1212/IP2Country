package service

import (
	"fmt"

	"github.com/evyataryagoni/ip2country/internal/logger"
	"github.com/evyataryagoni/ip2country/internal/metrics"
	"github.com/evyataryagoni/ip2country/internal/models"
	"github.com/evyataryagoni/ip2country/internal/store"
	"github.com/go-playground/validator/v10"
)

// IPService handles business logic for IP lookups
// This is the service layer - it sits between handlers and stores
//
// Responsibilities:
//   - Validate input (IP format)
//   - Call the store
//   - Handle errors
//   - Transform data if needed
type IPService struct {
	store     store.Store          // The datastore (CSV, MySQL, or Redis)
	validator *validator.Validate  // Validator for input validation
	metrics   *metrics.Metrics     // Metrics collector
	logger    *logger.Logger       // Structured logger
}

// NewIPService creates a new IP service
// This is the constructor function (common Go pattern)
//
// Parameters:
//   - store: any implementation of the Store interface
//   - m: metrics collector (optional, can be nil)
//   - log: logger (optional, can be nil)
//
// Returns:
//   - *IPService: pointer to the created service
func NewIPService(store store.Store, m *metrics.Metrics, log *logger.Logger) *IPService {
	if log == nil {
		log = logger.NewDefault()
	}
	return &IPService{
		store:     store,
		validator: validator.New(), // Create a new validator instance
		metrics:   m,
		logger:    log.WithComponent("IPService"),
	}
}

// LookupIP looks up geographic information for an IP address
// This is the main business logic method
//
// Flow:
//   1. Validate IP format
//   2. Query the store
//   3. Return result or error
//
// Parameters:
//   - ip: the IP address to lookup (IPv4 or IPv6)
//
// Returns:
//   - *models.IPLocation: the location data
//   - error: validation error, not found error, or store error
func (s *IPService) LookupIP(ip string) (*models.IPLocation, error) {
	// Step 1: Validate IP format
	// The validator uses struct tags, but we can also validate individual values
	// "ip" is a built-in validation tag that checks for valid IPv4/IPv6
	err := s.validator.Var(ip, "required,ip")
	if err != nil {
		// Log and track validation error
		s.logger.Warn().Str("ip", ip).Msg("Invalid IP address format")
		if s.metrics != nil {
			s.metrics.IPLookupsErrors.WithLabelValues("validation").Inc()
		}
		// Return a user-friendly error message
		return nil, fmt.Errorf("invalid IP address format")
	}

	// Step 2: Query the store
	// The store handles the actual data access (CSV, MySQL, Redis)
	s.logger.Debug().Str("ip", ip).Msg("Looking up IP address")
	location, err := s.store.FindByIP(ip)
	if err != nil {
		// Track not found or error
		if s.metrics != nil {
			if err.Error() == "IP address not found" {
				s.logger.Debug().Str("ip", ip).Msg("IP address not found")
				s.metrics.IPLookupsNotFound.Inc()
				s.metrics.IPLookupsTotal.WithLabelValues("not_found").Inc()
			} else {
				s.logger.Error().Err(err).Str("ip", ip).Msg("Store error during IP lookup")
				s.metrics.IPLookupsErrors.WithLabelValues("store_error").Inc()
			}
		}
		// Return the error from the store (usually "not found")
		return nil, err
	}

	// Step 3: Return the result
	// Track successful lookup
	s.logger.Info().
		Str("ip", ip).
		Str("city", location.City).
		Str("country", location.Country).
		Msg("IP lookup successful")
	if s.metrics != nil {
		s.metrics.IPLookupsTotal.WithLabelValues("success").Inc()
	}
	// No transformation needed - just pass through the location data
	return location, nil
}

// Close cleans up resources
// Should be called when the service is no longer needed
// This will close the underlying store (database connections, etc.)
func (s *IPService) Close() error {
	return s.store.Close()
}
