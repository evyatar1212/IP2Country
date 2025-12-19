package v1

import (
	"github.com/evyataryagoni/ip2country/internal/handler"
	"github.com/go-chi/chi/v5"
)

// SetupRoutes configures all v1 API routes
func SetupRoutes(ipHandler *handler.IPHandler) chi.Router {
	r := chi.NewRouter()

	r.Get("/find-country", ipHandler.FindCountry)

	// Future v1 endpoints can be added here:
	// r.Get("/lookup", ipHandler.Lookup)
	// r.Get("/batch", ipHandler.BatchLookup)

	return r
}
