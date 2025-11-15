.PHONY: build run test clean deps fmt banking fraud guardrail clearance scoring

# Build the application
build:
	@echo "Building Agent..."
	@go build -o bin/agent cmd/server/main.go

# Run the application (set AGENT_TYPE env var)
run:
	@echo "Running Agent..."
	@go run cmd/server/main.go

# Run Banking Agent
banking:
	@echo "Running Banking Agent..."
	@AGENT_TYPE=BANKING SERVER_PORT=8001 AGENT_ENDPOINT=http://localhost:8001 go run cmd/server/main.go

# Run Fraud Agent
fraud:
	@echo "Running Fraud Agent..."
	@AGENT_TYPE=FRAUD SERVER_PORT=8002 AGENT_ENDPOINT=http://localhost:8002 go run cmd/server/main.go

# Run Guardrail Agent
guardrail:
	@echo "Running Guardrail Agent..."
	@AGENT_TYPE=GUARDRAIL SERVER_PORT=8003 AGENT_ENDPOINT=http://localhost:8003 go run cmd/server/main.go

# Run Clearance Agent
clearance:
	@echo "Running Clearance Agent..."
	@AGENT_TYPE=CLEARANCE SERVER_PORT=8004 AGENT_ENDPOINT=http://localhost:8004 go run cmd/server/main.go

# Run Scoring Agent
scoring:
	@echo "Running Scoring Agent..."
	@AGENT_TYPE=SCORING SERVER_PORT=8005 AGENT_ENDPOINT=http://localhost:8005 go run cmd/server/main.go

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

