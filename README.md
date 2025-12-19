# IP2Country Service

A high-performance Go service that provides IP geolocation lookups with multiple storage backends, rate limiting, and comprehensive test coverage.

## Features

- **Multiple Storage Backends**: CSV, Redis, MySQL
- **Rate Limiting**: Per-IP rate limiting with in-memory or Redis backends
- **RESTful API**: Clean HTTP API with proper status codes
- **API Documentation**: Interactive Swagger/OpenAPI documentation
- **Observability**: Prometheus metrics, structured logging (zerolog)
- **Docker Support**: Full Docker and docker-compose setup
- **Comprehensive Tests**: 71 unit tests with 60-95% coverage
- **Production Ready**: Middleware chain, graceful shutdown, connection pooling

## API Endpoints

### Find Country by IP
```http
GET /v1/find-country?ip=8.8.8.8
```

**Response (200 OK):**
```json
{
  "city": "Mountain View",
  "country": "United States"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid IP format or missing parameter
- `404 Not Found` - IP not in database
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

### Health Check
```http
GET /health
```

Returns `200 OK` if the service is running.

### Prometheus Metrics
```http
GET /metrics
```

Returns Prometheus-formatted metrics including:
- Request counts and durations
- Rate limiting metrics
- Datastore performance
- Error rates

### API Documentation (Swagger UI)
```http
GET /swagger/index.html
```

Interactive API documentation with examples and try-it-out functionality.

## Quick Start

### Prerequisites
- Go 1.21+ (for local development)
- Docker & Docker Compose (for containerized deployment)
- Redis (optional, for Redis-based rate limiting or storage)
- MySQL (optional, for MySQL storage)

### Local Development

```bash
# 1. Clone the repository
git clone <repository-url>
cd ip2country

# 2. Install dependencies
go mod download

# 3. Install Swagger CLI (for API documentation)
go install github.com/swaggo/swag/cmd/swag@latest

# 4. Generate Swagger documentation
swag init -g cmd/server/main.go -o docs

# 5. Run tests
go test ./internal/... -cover

# 6. Run the service (uses CSV by default)
go run cmd/server/main.go
```

The service will start on `http://localhost:3000`:
- API: http://localhost:3000/v1/find-country?ip=8.8.8.8
- Health: http://localhost:3000/health
- Metrics: http://localhost:3000/metrics
- Swagger: http://localhost:3000/swagger/index.html

### Docker

```bash
# Start all services (app + MySQL + Redis)
docker-compose up

# Test the API
curl "http://localhost:3000/v1/find-country?ip=8.8.8.8"

# View logs
docker-compose logs -f app

# Stop services
docker-compose down
```

### Docker with Tests

```bash
# Build with tests (fails if tests don't pass)
docker build -f Dockerfile.test -t ip2country:test .

# Run the tested image
docker run -p 3000:3000 ip2country:test
```

## Configuration

Configure via environment variables or `.env` file:

```bash
# Server Configuration
PORT=3000                 # Server port (default: 3000)

# Rate Limiting
RATE_LIMITER_TYPE=memory  # "memory" or "redis"
RATE_LIMIT=10             # Number of requests allowed
RATE_LIMIT_WINDOW=1       # Time window in seconds

# Data Store
DATASTORE_TYPE=csv        # "csv", "redis", or "mysql"
DATASTORE_PATH=./data/ip2country.csv  # Path to CSV file

# Redis Configuration (if using Redis store or limiter)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=          # Leave empty if no password
REDIS_DB=0               # Redis database number (0-15)

# MySQL Configuration (if using MySQL store)
MYSQL_DSN=root:password@tcp(localhost:3306)/ip2country?parseTime=true
```

### Configuration Examples

#### Example 1: CSV Store + Memory Rate Limiter (Default)
```bash
# Simplest setup - no external dependencies
DATASTORE_TYPE=csv
DATASTORE_PATH=./data/ip2country.csv
RATE_LIMITER_TYPE=memory
RATE_LIMIT=100
RATE_LIMIT_WINDOW=5  # 100 requests per 5 seconds
```

#### Example 2: Redis Store + Redis Rate Limiter (Production)
```bash
# Best for multi-server deployments
DATASTORE_TYPE=redis
RATE_LIMITER_TYPE=redis
REDIS_ADDR=redis:6379
REDIS_PASSWORD=your_password
REDIS_DB=0
RATE_LIMIT=1000
RATE_LIMIT_WINDOW=60  # 1000 requests per minute
```

#### Example 3: MySQL Store + Memory Rate Limiter
```bash
# Good for persistent storage with SQL queries
DATASTORE_TYPE=mysql
MYSQL_DSN=user:password@tcp(mysql:3306)/ip2country?parseTime=true
RATE_LIMITER_TYPE=memory
RATE_LIMIT=50
RATE_LIMIT_WINDOW=1  # 50 requests per second
```

### Datastore Options

#### 1. CSV Store (Default)
**Best for:** Development, single-server deployments, small datasets

```bash
DATASTORE_TYPE=csv
DATASTORE_PATH=./data/ip2country.csv
```

**Pros:**
- No external dependencies
- Fast lookups (~250ns)
- Simple setup

**Cons:**
- Requires server restart to update data
- Entire dataset loaded in memory

#### 2. Redis Store
**Best for:** Production, distributed systems, frequent updates

```bash
DATASTORE_TYPE=redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

**Pros:**
- Fast lookups (~1-2ms)
- Horizontal scaling
- Data can be updated without restart
- Shared across multiple server instances

**Cons:**
- Requires Redis server
- Network latency

**Load data into Redis:**
```bash
# First time setup or data refresh
go run cmd/load-redis/main.go
```

The service will auto-load sample data if Redis is empty on startup.

#### 3. MySQL Store
**Best for:** Enterprise, complex queries, persistent storage

```bash
DATASTORE_TYPE=mysql
MYSQL_DSN=root:password@tcp(localhost:3306)/ip2country?parseTime=true
```

**Pros:**
- ACID compliance
- Complex queries and joins
- Persistent storage
- Backup and replication

**Cons:**
- Slower than in-memory (~2-5ms)
- Requires MySQL server

### Rate Limiting Options

#### 1. Memory Rate Limiter (Default)
**Best for:** Single-server deployments, development

```bash
RATE_LIMITER_TYPE=memory
RATE_LIMIT=100
RATE_LIMIT_WINDOW=5  # 100 requests per 5 seconds per IP
```

**Pros:**
- No external dependencies
- Very fast (~250ns)
- Simple setup

**Cons:**
- Not shared across servers
- State lost on restart

#### 2. Redis Rate Limiter
**Best for:** Multi-server deployments, distributed systems

```bash
RATE_LIMITER_TYPE=redis
RATE_LIMIT=1000
RATE_LIMIT_WINDOW=60  # 1000 requests per minute per IP
REDIS_ADDR=localhost:6379
```

**Pros:**
- Shared across all servers
- Accurate rate limiting in distributed systems
- Persistent across restarts

**Cons:**
- Requires Redis server
- Slightly slower (~1ms)

**Rate Limit Calculation:**
- The effective rate is: `RATE_LIMIT / RATE_LIMIT_WINDOW` requests per second
- Example: `RATE_LIMIT=100` and `RATE_LIMIT_WINDOW=5` = 20 req/s
- Fractional rates supported: `RATE_LIMIT=1` and `RATE_LIMIT_WINDOW=5` = 0.2 req/s (1 request per 5 seconds)

## Architecture

The service follows **Clean Architecture** / **Hexagonal Architecture** principles:

```
┌─────────────────────────────────────────────────────────┐
│                      HTTP Layer                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Router → Middleware → Handler                   │  │
│  │  (Chi)    (Logging,      (HTTP)                  │  │
│  │           RateLimit,                             │  │
│  │           Metrics)                               │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│                   Business Logic Layer                  │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Service (IP validation, error handling)         │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│                    Data Access Layer                    │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Store Interface                                 │  │
│  │    ├─ CSV Store                                  │  │
│  │    ├─ Redis Store                                │  │
│  │    └─ MySQL Store                                │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Project Structure

```
├── cmd/
│   ├── server/             # Main application entry point
│   └── load-redis/         # Redis data loading tool
├── internal/
│   ├── handler/            # HTTP handlers (94.7% coverage)
│   ├── service/            # Business logic (68.0% coverage)
│   ├── store/              # Data access layer (60.4% coverage)
│   │   ├── store.go        # Interface definition
│   │   ├── csv_store.go    # In-memory CSV implementation
│   │   ├── redis_store.go  # Redis implementation
│   │   └── mysql_store.go  # MySQL implementation
│   ├── middleware/         # HTTP middleware
│   │   ├── rate_limit.go   # Rate limiting middleware
│   │   ├── logging.go      # Structured logging middleware
│   │   └── metrics.go      # Prometheus metrics middleware
│   ├── limiter/            # Rate limiting implementations
│   │   ├── limiter.go      # Interface + token bucket algorithm
│   │   ├── rate_limiter.go # In-memory implementation
│   │   ├── redis_limiter.go# Distributed Redis implementation
│   │   └── factory.go      # Factory pattern for limiter creation
│   ├── router/             # Route configuration
│   │   ├── router.go       # Main router setup
│   │   └── v1/routes.go    # API v1 routes
│   ├── config/             # Configuration management
│   ├── logger/             # Structured logging (zerolog)
│   ├── metrics/            # Prometheus metrics definitions
│   └── models/             # Data models
├── data/                   # CSV data files
├── docs/                   # Swagger documentation (auto-generated)
└── docker-compose.yml      # Full stack setup
```

## Testing

### Run All Tests
```bash
go test ./internal/... -v
```

### Run with Coverage
```bash
go test ./internal/... -cover

# Generate HTML coverage report
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```

### Test Coverage

| Component | Coverage | Tests |
|-----------|----------|-------|
| Handler Layer | 94.7% | 13 tests |
| Service Layer | 68.0% | 13 tests |
| Store Layer | 60.4% | 37 tests |
| Middleware | 30.4% | 10 tests |
| **Total** | **71 tests** | ✅ All passing |

### Test Structure

- **Unit Tests**: Test individual components in isolation
- **Mock Implementations**: `mock_store.go`, `mock_limiter.go`
- **Table-Driven Tests**: Multiple scenarios per test function
- **Integration Tests**: Docker-based testing available

**Technologies Used:**
- Standard Go `testing` package
- `github.com/alicebob/miniredis/v2` - In-memory Redis for testing
- `github.com/DATA-DOG/go-sqlmock` - MySQL mock for testing

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── handler/
│   │   ├── ip_handler.go        # HTTP handlers
│   │   └── ip_handler_test.go
│   ├── service/
│   │   ├── ip_service.go        # Business logic
│   │   └── ip_service_test.go
│   ├── store/
│   │   ├── store.go             # Store interface
│   │   ├── csv_store.go         # CSV implementation
│   │   ├── csv_store_test.go
│   │   ├── redis_store.go       # Redis implementation
│   │   ├── redis_store_test.go
│   │   ├── mysql_store.go       # MySQL implementation
│   │   ├── mysql_store_test.go
│   │   └── mock_store.go        # Test mock
│   ├── middleware/
│   │   ├── rate_limit.go
│   │   ├── rate_limit_test.go
│   │   ├── logging.go
│   │   └── metrics.go
│   └── limiter/
│       ├── rate_limiter.go      # In-memory limiter
│       ├── redis_limiter.go     # Distributed limiter
│       ├── limiter_test.go
│       └── mock_limiter.go      # Test mock
├── data/
│   └── ip2country.csv           # IP database
├── Dockerfile                    # Development Dockerfile
├── Dockerfile.test               # Production Dockerfile with tests
├── docker-compose.yml            # Full stack setup
├── go.mod
├── go.sum
└── README.md
```

## Dependencies

### Direct Dependencies
```go
// Web Framework
github.com/go-chi/chi/v5           // Lightweight, composable router

// Validation
github.com/go-playground/validator/v10  // Struct validation

// Database
github.com/redis/go-redis/v9       // Redis client
gorm.io/gorm                       // ORM for MySQL
gorm.io/driver/mysql               // MySQL driver for GORM

// Logging & Metrics
github.com/rs/zerolog              // Structured logging
github.com/prometheus/client_golang // Prometheus metrics

// Configuration
github.com/joho/godotenv           // .env file support

// API Documentation
github.com/swaggo/swag             // Swagger/OpenAPI generator
github.com/swaggo/http-swagger/v2  // Swagger UI for Chi
```

### Testing Dependencies
```go
github.com/alicebob/miniredis/v2   // In-memory Redis for testing
github.com/DATA-DOG/go-sqlmock     // SQL mock for testing
```

## Performance

- **In-Memory Store (CSV)**: ~250ns per lookup
- **Redis Store**: ~1-2ms per lookup (with network)
- **MySQL Store**: ~2-5ms per lookup (with network)
- **Rate Limiter**: ~250ns (in-memory), ~1ms (Redis)

## Design Decisions

### Why Clean Architecture?
- **Testability**: Each layer can be tested independently
- **Flexibility**: Easy to swap implementations (CSV → Redis → MySQL)
- **Maintainability**: Clear separation of concerns

### Why Multiple Storage Backends?
- **CSV**: Simple, fast, no dependencies (development)
- **Redis**: Fast, distributed, scales horizontally (production)
- **MySQL**: Persistent, ACID compliant, complex queries (enterprise)

### Why Rate Limiting?
- **Protection**: Prevents abuse and DoS attacks
- **Fairness**: Ensures equal access for all users
- **Scalability**: Distributed rate limiting via Redis

## Observability

### Structured Logging
The service uses **zerolog** for structured, high-performance logging:

```json
{
  "level": "info",
  "component": "IPService",
  "ip": "8.8.8.8",
  "city": "Mountain View",
  "country": "United States",
  "time": "2025-01-01T12:00:00Z",
  "message": "IP lookup successful"
}
```

### Prometheus Metrics

Available at `/metrics`:

**Request Metrics:**
- `http_requests_total` - Total HTTP requests (by method, endpoint, status)
- `http_request_duration_seconds` - Request latency histogram
- `http_request_size_bytes` - Request size histogram
- `http_response_size_bytes` - Response size histogram

**Application Metrics:**
- `ip_lookups_total` - Total IP lookups (by result: success/not_found)
- `ip_lookups_not_found_total` - Total not found lookups
- `ip_lookups_errors_total` - Total lookup errors (by error_type)

**Datastore Metrics:**
- `datastore_queries_total` - Total datastore queries
- `datastore_query_duration_seconds` - Query latency
- `datastore_cache_hits_total` - Cache hits vs misses
- `datastore_connections_open` - Open database connections

## Production Considerations

✅ **Implemented**
- Structured logging with zerolog
- Prometheus metrics for observability
- Rate limiting (in-memory or distributed via Redis)
- Health check endpoint for load balancers
- Graceful shutdown handling
- Database connection pooling
- Comprehensive error handling
- Docker support with multi-stage builds
- API documentation with Swagger/OpenAPI
- Clean architecture for maintainability
- Extensive test coverage (71 tests)

📋 **Future Enhancements**
- Authentication/API keys
- Request/response caching layer
- Geolocation by coordinates (reverse lookup)
- Bulk IP lookup endpoint (`POST /v1/batch`)
- Admin API for data management
- CircuitBreaker for external dependencies
- Distributed tracing (OpenTelemetry)
- Rate limiting by API key (not just IP)

## License

This project was created as a home assignment demonstration.

## Author

Evyatar Yagoni
