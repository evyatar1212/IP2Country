package v1

import (
	"github.com/evyataryagoni/ip2country/internal/handler"
	"github.com/go-chi/chi/v5"
)

// SetupRoutes configures all v1 API routes
// This function is called by the main router to setup /v1/* endpoints
//
// Parameters:
//   - ipHandler: the IP lookup handler
//
// Returns:
//   - chi.Router: configured v1 router
func SetupRoutes(ipHandler *handler.IPHandler) chi.Router {
	r := chi.NewRouter()

	// IP lookup endpoint
	// GET /v1/find-country?ip=<ip>
	r.Get("/find-country", ipHandler.FindCountry)

	// Future v1 endpoints can be added here:
	// r.Get("/lookup", ipHandler.Lookup)
	// r.Get("/batch", ipHandler.BatchLookup)

	return r
}
