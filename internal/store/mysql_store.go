package store

import (
	"fmt"

	"github.com/evyataryagoni/ip2country/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// IPCountryModel is the GORM model for the ip2country table
// GORM uses struct tags to map to database columns
type IPCountryModel struct {
	IP      string `gorm:"column:ip;primaryKey"` // Primary key
	City    string `gorm:"column:city"`
	Country string `gorm:"column:country"`
}

// TableName specifies the table name for GORM
// By default, GORM would pluralize to "ip_country_models"
// This override tells GORM to use "ip2country" instead
func (IPCountryModel) TableName() string {
	return "ip2country"
}

// MySQLStore implements Store interface using MySQL with GORM
// GORM provides ORM features like automatic query building and connection pooling
type MySQLStore struct {
	db *gorm.DB // GORM database instance
}

// NewMySQLStore creates a new MySQL store using GORM
//
// Parameters:
//   - dsn: Data Source Name (connection string)
//     Format: user:password@tcp(host:port)/dbname?parseTime=true
//     Example: root:password@tcp(localhost:3306)/ip2country?parseTime=true
//
// Returns:
//   - *MySQLStore: pointer to the created store
//   - error: any error that occurred during connection
func NewMySQLStore(dsn string) (*MySQLStore, error) {
	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Disable query logging (set to Info for debugging)
	}

	// Open database connection with GORM
	// GORM handles connection pooling automatically
	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL with GORM: %w", err)
	}

	// Get underlying SQL database for configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)   // Maximum number of open connections
	sqlDB.SetMaxIdleConns(5)    // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(300) // Maximum connection lifetime (5 minutes)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	return &MySQLStore{db: db}, nil
}

// FindByIP looks up an IP address using GORM
// Implements the Store interface method
//
// GORM automatically generates the SQL query based on the model
func (s *MySQLStore) FindByIP(ip string) (*models.IPLocation, error) {
	var record IPCountryModel

	// GORM query: SELECT * FROM ip2country WHERE ip = ? LIMIT 1
	// First() finds the first record matching the condition
	result := s.db.Where("ip = ?", ip).First(&record)

	// Check for errors
	if result.Error != nil {
		// GORM returns gorm.ErrRecordNotFound when no rows found
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("IP address not found")
		}
		// Other database errors
		return nil, fmt.Errorf("database query failed: %w", result.Error)
	}

	// Convert GORM model to our domain model
	return &models.IPLocation{
		IP:      record.IP,
		City:    record.City,
		Country: record.Country,
	}, nil
}

// Close closes the database connection
// Should be called when the application shuts down
func (s *MySQLStore) Close() error {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
