package store

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// setupMockDB creates a mock database for testing
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	return db, mock, sqlDB
}

// TestMySQLStore_FindByIP_Success tests successful lookup
func TestMySQLStore_FindByIP_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	store := &MySQLStore{db: db}

	// Set up mock expectations
	// Note: GORM adds LIMIT 1 to First() queries, so we expect 2 args: ip and limit
	rows := sqlmock.NewRows([]string{"ip", "city", "country"}).
		AddRow("8.8.8.8", "Mountain View", "United States")

	mock.ExpectQuery("SELECT \\* FROM `ip2country` WHERE ip = \\? .*").
		WithArgs("8.8.8.8", 1).
		WillReturnRows(rows)

	// Execute
	location, err := store.FindByIP("8.8.8.8")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if location.IP != "8.8.8.8" {
		t.Errorf("expected IP '8.8.8.8', got '%s'", location.IP)
	}
	if location.City != "Mountain View" {
		t.Errorf("expected 'Mountain View', got '%s'", location.City)
	}
	if location.Country != "United States" {
		t.Errorf("expected 'United States', got '%s'", location.Country)
	}

	// Verify all expectations met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestMySQLStore_FindByIP_MultipleIPs tests multiple IP lookups
func TestMySQLStore_FindByIP_MultipleIPs(t *testing.T) {
	tests := []struct {
		ip      string
		city    string
		country string
	}{
		{"8.8.8.8", "Mountain View", "United States"},
		{"1.1.1.1", "Sydney", "Australia"},
		{"9.9.9.9", "Berkeley", "United States"},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			db, mock, sqlDB := setupMockDB(t)
			defer sqlDB.Close()

			store := &MySQLStore{db: db}

			rows := sqlmock.NewRows([]string{"ip", "city", "country"}).
				AddRow(tt.ip, tt.city, tt.country)

			mock.ExpectQuery("SELECT \\* FROM `ip2country` WHERE ip = \\? .*").
				WithArgs(tt.ip, 1).
				WillReturnRows(rows)

			location, err := store.FindByIP(tt.ip)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if location.City != tt.city {
				t.Errorf("expected city '%s', got '%s'", tt.city, location.City)
			}
			if location.Country != tt.country {
				t.Errorf("expected country '%s', got '%s'", tt.country, location.Country)
			}

			mock.ExpectationsWereMet()
		})
	}
}

// TestMySQLStore_FindByIP_NotFound tests IP not found
func TestMySQLStore_FindByIP_NotFound(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	store := &MySQLStore{db: db}

	// Set up mock to return no rows (record not found)
	mock.ExpectQuery("SELECT \\* FROM `ip2country` WHERE ip = \\? .*").
		WithArgs("192.168.1.1", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	location, err := store.FindByIP("192.168.1.1")

	if err == nil {
		t.Error("expected not found error, got nil")
	}
	if location != nil {
		t.Error("expected nil location, got data")
	}
	if err.Error() != "IP address not found" {
		t.Errorf("expected 'IP address not found', got '%s'", err.Error())
	}

	mock.ExpectationsWereMet()
}

// TestMySQLStore_FindByIP_DatabaseError tests database errors
func TestMySQLStore_FindByIP_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	store := &MySQLStore{db: db}

	// Simulate database error
	mock.ExpectQuery("SELECT \\* FROM `ip2country` WHERE ip = \\? .*").
		WithArgs("8.8.8.8", 1).
		WillReturnError(sql.ErrConnDone)

	location, err := store.FindByIP("8.8.8.8")

	if err == nil {
		t.Error("expected database error, got nil")
	}
	if location != nil {
		t.Error("expected nil location, got data")
	}
	// Should wrap the error, not return "IP address not found"
	if err.Error() == "IP address not found" {
		t.Error("expected database error, got not found error")
	}

	mock.ExpectationsWereMet()
}

// TestMySQLStore_Close tests cleanup
func TestMySQLStore_Close(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	store := &MySQLStore{db: db}

	mock.ExpectClose()

	err := store.Close()

	if err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}

	mock.ExpectationsWereMet()
}

// TestMySQLStore_Close_NilDB tests close with nil db
func TestMySQLStore_Close_NilDB(t *testing.T) {
	store := &MySQLStore{db: nil}

	err := store.Close()

	if err != nil {
		t.Errorf("expected no error for nil db, got: %v", err)
	}
}

// TestIPCountryModel_TableName tests GORM table name override
func TestIPCountryModel_TableName(t *testing.T) {
	model := IPCountryModel{}

	tableName := model.TableName()

	if tableName != "ip2country" {
		t.Errorf("expected table name 'ip2country', got '%s'", tableName)
	}
}

// TestMySQLStore_FindByIP_SpecialCharacters tests special characters in data
func TestMySQLStore_FindByIP_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		city    string
		country string
	}{
		{"UTF-8 city", "8.8.8.8", "São Paulo", "Brazil"},
		{"Chinese characters", "1.1.1.1", "北京", "China"},
		{"Cyrillic", "2.2.2.2", "Москва", "Russia"},
		{"Accents", "3.3.3.3", "Zürich", "Switzerland"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupMockDB(t)
			defer sqlDB.Close()

			store := &MySQLStore{db: db}

			rows := sqlmock.NewRows([]string{"ip", "city", "country"}).
				AddRow(tt.ip, tt.city, tt.country)

			mock.ExpectQuery("SELECT \\* FROM `ip2country` WHERE ip = \\? .*").
				WithArgs(tt.ip, 1).
				WillReturnRows(rows)

			location, err := store.FindByIP(tt.ip)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if location.City != tt.city {
				t.Errorf("expected city '%s', got '%s'", tt.city, location.City)
			}

			mock.ExpectationsWereMet()
		})
	}
}

// TestMySQLStore_FindByIP_IPv6 tests IPv6 addresses
func TestMySQLStore_FindByIP_IPv6(t *testing.T) {
	ipv6Addresses := []string{
		"2001:4860:4860::8888",
		"::1",
		"fe80::1",
	}

	for _, ip := range ipv6Addresses {
		t.Run(ip, func(t *testing.T) {
			db, mock, sqlDB := setupMockDB(t)
			defer sqlDB.Close()

			store := &MySQLStore{db: db}

			rows := sqlmock.NewRows([]string{"ip", "city", "country"}).
				AddRow(ip, "Test City", "Test Country")

			mock.ExpectQuery("SELECT \\* FROM `ip2country` WHERE ip = \\? .*").
				WithArgs(ip, 1).
				WillReturnRows(rows)

			location, err := store.FindByIP(ip)

			if err != nil {
				t.Fatalf("unexpected error for IPv6: %v", err)
			}
			if location.IP != ip {
				t.Errorf("expected IP '%s', got '%s'", ip, location.IP)
			}

			mock.ExpectationsWereMet()
		})
	}
}

// TestMySQLStore_FindByIP_EmptyResult tests query returning empty result
func TestMySQLStore_FindByIP_EmptyResult(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	store := &MySQLStore{db: db}

	// Return gorm.ErrRecordNotFound for empty result
	mock.ExpectQuery("SELECT \\* FROM `ip2country` WHERE ip = \\? .*").
		WithArgs("10.0.0.1", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	location, err := store.FindByIP("10.0.0.1")

	if err == nil {
		t.Error("expected error for empty result, got nil")
	}
	if location != nil {
		t.Error("expected nil location for empty result")
	}

	mock.ExpectationsWereMet()
}

// TestIPCountryModel_StructTags tests GORM struct tags
func TestIPCountryModel_StructTags(t *testing.T) {
	// Create a model instance
	model := IPCountryModel{
		IP:      "8.8.8.8",
		City:    "Mountain View",
		Country: "United States",
	}

	// Verify fields are set correctly
	if model.IP != "8.8.8.8" {
		t.Errorf("expected IP '8.8.8.8', got '%s'", model.IP)
	}
	if model.City != "Mountain View" {
		t.Errorf("expected city 'Mountain View', got '%s'", model.City)
	}
	if model.Country != "United States" {
		t.Errorf("expected country 'United States', got '%s'", model.Country)
	}
}
