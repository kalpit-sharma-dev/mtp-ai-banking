.PHONY: build run test clean docker-build docker-run

# Build the application
build:
	@echo "Building MCP Server..."
	@go build -o bin/mcp-server cmd/server/main.go

# Run the application
run:
	@echo "Running MCP Server..."
	@go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./...

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t mcp-server:latest .

# Docker run
docker-run:
	@echo "Running Docker container..."
	@docker-compose up

