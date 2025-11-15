# AI Banking MCP Server

A Model Context Protocol (MCP) server built in Golang for orchestrating AI Banking operations using Agentic AI and ML models.

## Overview

This MCP Server acts as the central orchestration layer for an AI Banking Platform, managing:
- Task routing and execution
- Agent registration and discovery
- Session and context management
- Rule-based routing decisions
- Integration with specialized AI agents (Banking, Fraud, Guardrail, Clearance, Scoring)

## Architecture

```
Mobile Banking / Net Banking Apps
            |
            v
        API Gateway
            |
            v
      MCP Server (Golang)
   -----------------------------------
   |     AI Skin Orchestrator        |
   -----------------------------------
   |     |        |         |        |
   v     v        v         v        v
Banking Agent | Fraud Agent | Guardrail Agent |
Clearance Agent | Scoring Agent
```

## Features

- **REST API** for task submission and management
- **Session Management** with Redis-backed context storage
- **Agent Registry** for dynamic agent registration and discovery
- **Context Router** for intelligent task routing based on rules and intent
- **Rule Engine** for configurable routing logic
- **Security** with API key authentication and rate limiting
- **Observability** with structured logging

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── controller/              # HTTP controllers
│   ├── middleware/              # HTTP middleware (auth, logging, rate limit)
│   ├── model/                   # Data models
│   ├── router/                  # Route definitions
│   ├── service/                 # Business logic services
│   └── utils/                   # Utility functions
├── proto/                       # gRPC definitions (future)
├── go.mod
├── go.sum
├── .env.example
└── README.md
```

## Prerequisites

- Go 1.21 or higher
- Redis (for session and task storage)
- (Optional) PostgreSQL (for future persistent storage)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd mcp-server
```

2. Install dependencies:
```bash
go mod download
```

3. Copy environment file:
```bash
cp .env.example .env
```

4. Update `.env` with your configuration (especially Redis connection)

5. Start Redis (if not already running):
```bash
# Using Docker
docker run -d -p 6379:6379 redis:latest

# Or using local Redis
redis-server
```

6. Run the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Task Management

- `POST /api/v1/submit-task` - Submit a new banking task
- `GET /api/v1/get-result/{taskID}` - Get task result by ID

### Agent Management

- `POST /api/v1/register-agent` - Register a new agent
- `GET /api/v1/agent/{agentID}` - Get agent details
- `GET /api/v1/agents` - List all registered agents

### Session Management

- `POST /api/v1/create-session` - Create a new session
- `GET /api/v1/get-session/{sessionID}` - Get session details

### Rule Management

- `POST /api/v1/rules/upload` - Upload routing rules
- `GET /api/v1/rules` - Get all routing rules

### Health Checks

- `GET /health` - Health check endpoint
- `GET /ready` - Readiness check endpoint

## Example Usage

### 1. Submit a Task

```bash
curl -X POST http://localhost:8080/api/v1/submit-task \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "user_id": "U10001",
    "channel": "MB",
    "intent": "TRANSFER_NEFT",
    "data": {
      "amount": 50000,
      "to_account": "XXXX4321"
    }
  }'
```

Response:
```json
{
  "task_id": "task_abc123",
  "session_id": "sess_xyz789",
  "status": "PENDING",
  "message": "Task submitted successfully",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### 2. Get Task Result

```bash
curl -X GET http://localhost:8080/api/v1/get-result/task_abc123 \
  -H "X-API-Key: your-api-key"
```

Response:
```json
{
  "task_id": "task_abc123",
  "status": "COMPLETED",
  "result": {
    "status": "APPROVED",
    "message": "Transaction processed successfully"
  },
  "risk_score": 0.12,
  "explanation": "Transaction is within user limits and behavior pattern is normal.",
  "completed_at": "2024-01-15T10:30:05Z"
}
```

### 3. Register an Agent

```bash
curl -X POST http://localhost:8080/api/v1/register-agent \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "name": "Custom Banking Agent",
    "type": "BANKING",
    "endpoint": "http://localhost:8001",
    "capabilities": ["CHECK_BALANCE", "FUND_TRANSFER"]
  }'
```

## Configuration

All configuration is managed through environment variables (see `.env.example`):

- **Server**: Port, host, timeouts
- **Redis**: Connection details
- **Security**: API key header, JWT secret, rate limits
- **Logging**: Log level and format

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/mcp-server cmd/server/main.go
```

### Running with Docker

```bash
docker-compose up
```

## Next Steps

This is **Layer 1: MCP Server**. The following layers will be built:

1. ✅ **MCP Server** (Current)
2. ⏳ **AI Skin Orchestrator** (Enhanced orchestration logic)
3. ⏳ **Agent Mesh** (Individual agent implementations)
4. ⏳ **ML Models** (Fraud detection, scoring models)
5. ⏳ **Banking Integrations** (MB, NB, DWH connections)
6. ⏳ **Event Streaming** (Kafka/PubSub integration)
7. ⏳ **Observability** (OpenTelemetry, Prometheus, Grafana)

## License

[Your License Here]

## Contributing

[Contributing Guidelines]

