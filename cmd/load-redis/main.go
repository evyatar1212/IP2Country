package main

import (
	"fmt"
	"log"

	"github.com/evyataryagoni/ip2country/internal/config"
	"github.com/evyataryagoni/ip2country/internal/store"
)

// This tool loads IP data from CSV into Redis
// Usage: go run cmd/load-redis/main.go
func main() {
	fmt.Println("🔄 Loading IP data into Redis...")

	// Load configuration
	appConfig := config.Load()

	// Connect to Redis
	fmt.Printf("📡 Connecting to Redis at %s...\n", appConfig.RedisAddr)
	redisStore, err := store.NewRedisStore(appConfig.RedisAddr, appConfig.RedisPassword, appConfig.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisStore.Close()

	fmt.Println("✅ Connected to Redis")

	// Load data from CSV
	fmt.Printf("📁 Loading data from %s...\n", appConfig.DatastorePath)
	if err := redisStore.LoadFromCSV(appConfig.DatastorePath); err != nil {
		log.Fatalf("Failed to load CSV data: %v", err)
	}

	fmt.Println("✅ Data loaded successfully!")
	fmt.Println("\n💡 You can now start the server with DATASTORE_TYPE=redis")
}
