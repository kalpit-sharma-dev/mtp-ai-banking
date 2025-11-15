.PHONY: build run test clean deps fmt

# Build the application
build:
	@echo "Building AI Skin Orchestrator..."
	@go build -o bin/ai-skin-orchestrator cmd/server/main.go

# Run the application
run:
	@echo "Running AI Skin Orchestrator..."
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

