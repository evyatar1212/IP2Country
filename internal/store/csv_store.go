package store

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/evyataryagoni/ip2country/internal/models"
)

// CSVStore implements Store interface using a CSV file
// It loads all data into memory for fast lookups
type CSVStore struct {
	// data maps IP addresses to location information
	// map[string]*models.IPLocation means: key=IP, value=pointer to IPLocation
	data map[string]*models.IPLocation
}

// NewCSVStore creates a new CSV store by reading a CSV file
// Parameters:
//   - filePath: path to the CSV file
//
// Returns:
//   - *CSVStore: pointer to the created store
//   - error: any error that occurred during file reading
//
// CSV Format: ip,city,country
// Example: 8.8.8.8,Mountain View,United States
func NewCSVStore(filePath string) (*CSVStore, error) {
	// Open the CSV file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	// defer means: execute this at the end of the function
	// Ensures file is closed even if we return early due to an error
	defer file.Close()

	// Create a CSV reader
	// csv.Reader knows how to parse CSV format
	reader := csv.NewReader(file)

	// Read all records at once
	// records is a 2D slice: [][]string
	// Example: [["ip","city","country"], ["8.8.8.8","Mountain View","United States"]]
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	// Check if file is empty
	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Create the store with an empty map
	// make(map[string]*models.IPLocation) creates a new map
	store := &CSVStore{
		data: make(map[string]*models.IPLocation),
	}

	// Parse each record (skip the header row)
	// range is like "for each" in other languages
	// i is the index, record is the value
	for i, record := range records {
		// Skip header row (first row with column names)
		if i == 0 {
			continue
		}

		// Validate record has exactly 3 columns
		if len(record) != 3 {
			// Skip invalid records instead of failing
			// In production, you might want to log this
			continue
		}

		// Extract fields from the CSV record
		ip := record[0]
		city := record[1]
		country := record[2]

		// Store in map: key=IP, value=IPLocation
		store.data[ip] = &models.IPLocation{
			IP:      ip,
			City:    city,
			Country: country,
		}
	}

	return store, nil
}

// FindByIP looks up an IP address in the store
// Implements the Store interface method
func (s *CSVStore) FindByIP(ip string) (*models.IPLocation, error) {
	// Look up IP in the map
	// In Go, map[key] returns two values:
	//   1. The value (or nil if not found)
	//   2. A boolean indicating if the key exists
	location, exists := s.data[ip]
	if !exists {
		// Return nil and an error if IP not found
		return nil, fmt.Errorf("IP address not found")
	}

	// Return the location data
	return location, nil
}

// Close cleans up resources
// For CSV store, there's nothing to clean up (all data is in memory)
// But we need this method to satisfy the Store interface
func (s *CSVStore) Close() error {
	// No resources to clean up
	return nil
}
