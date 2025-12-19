package store

import "github.com/evyataryagoni/ip2country/internal/models"

// Store defines the interface for IP lookup operations
// Allows multiple implementations (CSV, MySQL, Redis) and easy testing with mocks
type Store interface {
	// FindByIP looks up geographic information for an IP address
	FindByIP(ip string) (*models.IPLocation, error)

	// Close cleans up resources (database connections, file handles, etc.)
	Close() error
}
