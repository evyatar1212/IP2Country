# Development Dockerfile with hot reload using Air
# This is for development only

# Use Go 1.25 to match your local version
FROM golang:1.25-alpine

# Install air for hot reload (new repository path)
RUN go install github.com/air-verse/air@latest

# Set working directory
WORKDIR /app

# Copy go mod files first (for caching)
# Note: go.sum might not exist yet if we have no dependencies
COPY go.mod ./
COPY go.sum* ./

# Download dependencies (if any)
RUN go mod download

# Copy the rest of the source code
# Note: In docker-compose, we'll mount the entire directory as a volume
# so changes are reflected immediately
COPY . .

# Expose port
EXPOSE 3000

# Run air for hot reload
CMD ["air", "-c", ".air.toml"]
