# Layer 1: MCP Server

The MCP (Model Context Protocol) Server is the central orchestration layer for the AI Banking Platform. It manages task routing, session management, agent coordination, and context tracking.

## Overview

The MCP Server acts as the universal orchestrator that:
- Manages context for every request
- Maintains session state
- Routes tasks to appropriate agents
- Provides structured responses
- Ensures security between agents
- Logs trace for each action

## Features

✅ **Task Management** - Submit, track, and retrieve tasks  
✅ **Session Management** - Redis-backed session storage  
✅ **Agent Registry** - Dynamic agent registration and discovery  
✅ **Context Routing** - Intelligent task routing based on rules  
✅ **Rule Engine** - Configurable routing rules  
✅ **REST API** - Complete REST API for all operations  
✅ **Security** - API key authentication and rate limiting  

## Installation

1. Navigate to the mcp-server directory:
```bash
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

4. Start Redis (required for session/task storage):
```bash
docker run -d -p 6379:6379 redis:latest
```

5. Run the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Task Management
- `POST /api/v1/submit-task` - Submit a banking task
- `GET /api/v1/get-result/{taskID}` - Get task result

### Agent Management
- `POST /api/v1/register-agent` - Register a new agent
- `GET /api/v1/agent/{agentID}` - Get agent details
- `GET /api/v1/agents` - List all agents

### Session Management
- `POST /api/v1/create-session` - Create a session
- `GET /api/v1/get-session/{sessionID}` - Get session details

### Rule Management
- `POST /api/v1/rules/upload` - Upload routing rules
- `GET /api/v1/rules` - Get all rules

### Health Checks
- `GET /health` - Health check
- `GET /ready` - Readiness check

## Example Usage

### Submit a Task

```bash
curl -X POST http://localhost:8080/api/v1/submit-task \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
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

### Get Task Result

```bash
curl -X GET http://localhost:8080/api/v1/get-result/task_abc123 \
  -H "X-API-Key: test-api-key"
```

## Configuration

See `.env.example` for configuration options:
- Server port and host
- Redis connection
- Security settings
- Logging configuration

## Architecture

The MCP Server consists of:
- **Session Manager** - Manages user sessions
- **Task Manager** - Handles task lifecycle
- **Agent Registry** - Manages agent registration
- **Context Router** - Routes tasks to agents
- **Rule Engine** - Evaluates routing rules
- **Orchestrator** - Coordinates task execution

## Integration

The MCP Server integrates with:
- **Layer 2**: AI Skin Orchestrator sends requests to MCP
- **Layer 3**: Agents register with and receive tasks from MCP
- **Layer 4**: Agents may call ML Models for predictions
- **Layer 5**: Agents may call Banking Integrations for data

## License

[Your License Here]

