package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/evyataryagoni/ip2country/internal/models"
	"github.com/redis/go-redis/v9"
)

// RedisStore implements Store interface using Redis
// Redis is an in-memory key-value store, perfect for fast lookups
type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStore creates a new Redis store
//
// Parameters:
//   - addr: Redis server address (e.g., "localhost:6379")
//   - password: Redis password (empty string if no password)
//   - db: Redis database number (0-15, default is 0)
//
// Returns:
//   - *RedisStore: pointer to the created store
//   - error: any error that occurred during connection
func NewRedisStore(addr, password string, db int) (*RedisStore, error) {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// Test the connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStore{
		client: client,
		ctx:    ctx,
	}, nil
}

// FindByIP looks up an IP address in Redis
// Implements the Store interface method
//
// Redis Key Format: ip:<ip_address>
// Example: ip:8.8.8.8
// Value: JSON-encoded IPLocation
func (s *RedisStore) FindByIP(ip string) (*models.IPLocation, error) {
	// Build Redis key
	key := fmt.Sprintf("ip:%s", ip)

	// Get value from Redis
	val, err := s.client.Get(s.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Key does not exist
			return nil, fmt.Errorf("IP address not found")
		}
		// Other Redis errors
		return nil, fmt.Errorf("Redis query failed: %w", err)
	}

	// Decode JSON
	var location models.IPLocation
	if err := json.Unmarshal([]byte(val), &location); err != nil {
		return nil, fmt.Errorf("failed to decode IP location: %w", err)
	}

	// IP field has json:"-" tag, so it's not in JSON - set it manually
	location.IP = ip

	return &location, nil
}

// Set adds or updates an IP address in Redis
// This is a helper method for populating Redis with data
//
// Parameters:
//   - ip: the IP address
//   - city: the city name
//   - country: the country name
func (s *RedisStore) Set(ip, city, country string) error {
	location := models.IPLocation{
		IP:      ip,
		City:    city,
		Country: country,
	}

	// Encode to JSON
	data, err := json.Marshal(location)
	if err != nil {
		return fmt.Errorf("failed to encode IP location: %w", err)
	}

	// Build Redis key
	key := fmt.Sprintf("ip:%s", ip)

	// Store in Redis (no expiration)
	if err := s.client.Set(s.ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to store in Redis: %w", err)
	}

	return nil
}

// LoadFromCSV loads data from a CSV file into Redis
// This is useful for initial data population
func (s *RedisStore) LoadFromCSV(csvPath string) error {
	// Create a temporary CSV store to read the data
	csvStore, err := NewCSVStore(csvPath)
	if err != nil {
		return fmt.Errorf("failed to load CSV: %w", err)
	}
	defer csvStore.Close()

	// Iterate through all IPs in the CSV store and add to Redis
	count := 0
	for ip, location := range csvStore.data {
		if err := s.Set(ip, location.City, location.Country); err != nil {
			return fmt.Errorf("failed to store IP %s: %w", ip, err)
		}
		count++
	}

	fmt.Printf("Loaded %d IP records into Redis\n", count)
	return nil
}

// IsEmpty checks if Redis has any IP data
// Returns true if no keys with "ip:" prefix exist
func (s *RedisStore) IsEmpty() (bool, error) {
	// Check if any keys with "ip:" prefix exist
	keys, err := s.client.Keys(s.ctx, "ip:*").Result()
	if err != nil {
		return false, fmt.Errorf("failed to check Redis keys: %w", err)
	}
	return len(keys) == 0, nil
}

// Close closes the Redis connection
// Should be called when the application shuts down
func (s *RedisStore) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
