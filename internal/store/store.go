package store

import "github.com/evyataryagoni/ip2country/internal/models"

// Store defines the interface for IP lookup operations
// This is a Go interface - it defines behavior without implementation
// Any type that implements these methods satisfies this interface
//
// Why use an interface?
// - Allows multiple implementations (CSV, MySQL, Redis)
// - Easy to test (can create mock stores)
// - Follows the Strategy design pattern
type Store interface {
	// FindByIP looks up geographic information for an IP address
	// Returns:
	//   - *models.IPLocation: pointer to location data (nil if not found)
	//   - error: any error that occurred during lookup
	FindByIP(ip string) (*models.IPLocation, error)

	// Close cleans up resources (database connections, file handles, etc.)
	// Should be called when the store is no longer needed
	// Returns error if cleanup fails
	Close() error
}
