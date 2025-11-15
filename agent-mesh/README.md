# Layer 3: Agent Mesh

The Agent Mesh consists of individual, independently deployable agents that handle specific banking operations. Each agent is a separate service that can be scaled and deployed independently.

## Agents

### 1. Banking Agent
Handles core banking operations:
- Fund transfers (NEFT, RTGS, IMPS, UPI)
- Balance checks
- Account statements
- Beneficiary management

**Port**: 8001 (default)

### 2. Fraud Agent
Performs fraud detection using ML models and pattern analysis:
- Transaction fraud scoring
- Device anomaly detection
- Location anomaly detection
- Velocity checks
- Behavioral pattern analysis

**Port**: 8002 (default)

### 3. Guardrail Agent
Validates RBI regulations and bank policies:
- Daily transaction limits
- Single transaction limits
- Velocity limits
- Beneficiary age validation
- KYC status checks
- RBI blacklist checks

**Port**: 8003 (default)

### 4. Clearance Agent
Handles loan approval and clearance decisions:
- Loan eligibility assessment
- Credit score evaluation
- Income-to-loan ratio checks
- Interest rate calculation
- Auto/manual clearance decisions

**Port**: 8004 (default)

### 5. Scoring Agent
Provides credit, fraud, and risk scoring:
- Credit score calculation
- Fraud risk scoring
- Overall risk assessment
- Risk categorization

**Port**: 8005 (default)

## Architecture

```
MCP Server (Layer 1)
        |
        v
Agent Mesh (Layer 3)
  ├── Banking Agent (Port 8001)
  ├── Fraud Agent (Port 8002)
  ├── Guardrail Agent (Port 8003)
  ├── Clearance Agent (Port 8004)
  └── Scoring Agent (Port 8005)
```

## Features

✅ **Independent Deployment** - Each agent runs as a separate service  
✅ **Auto-Registration** - Agents automatically register with MCP Server  
✅ **Health Checks** - Each agent exposes `/health` endpoint  
✅ **Scalable** - Agents can be scaled independently  
✅ **Type-Safe** - Each agent implements specific business logic  
✅ **REST API** - Standard REST endpoints for all agents  

## Installation

1. Navigate to the agent-mesh directory:
```bash
cd agent-mesh
```

2. Install dependencies:
```bash
go mod download
```

3. Copy environment file:
```bash
cp .env.example .env
```

4. Configure agent type in `.env`:
```bash
AGENT_TYPE=BANKING  # or FRAUD, GUARDRAIL, CLEARANCE, SCORING
SERVER_PORT=8001    # Use different ports for different agents
AGENT_ENDPOINT=http://localhost:8001
```

5. Run the agent:
```bash
go run cmd/server/main.go
```

## Running Multiple Agents

Each agent runs independently. To run multiple agents:

### Terminal 1 - Banking Agent
```bash
cd agent-mesh
export AGENT_TYPE=BANKING
export SERVER_PORT=8001
export AGENT_ENDPOINT=http://localhost:8001
go run cmd/server/main.go
```

### Terminal 2 - Fraud Agent
```bash
cd agent-mesh
export AGENT_TYPE=FRAUD
export SERVER_PORT=8002
export AGENT_ENDPOINT=http://localhost:8002
go run cmd/server/main.go
```

### Terminal 3 - Guardrail Agent
```bash
cd agent-mesh
export AGENT_TYPE=GUARDRAIL
export SERVER_PORT=8003
export AGENT_ENDPOINT=http://localhost:8003
go run cmd/server/main.go
```

And so on for other agents.

## API Endpoints

All agents expose the same API:

### Process Request

**POST** `/api/v1/process`

Processes a request through the agent.

**Request:**
```json
{
  "agent_id": "BANKING",
  "task": "TRANSFER_NEFT",
  "input_context": {
    "user_id": "U10001",
    "data": {
      "amount": 50000,
      "to_account": "XXXX4321"
    }
  },
  "session_id": "sess_abc123",
  "request_id": "req_xyz789"
}
```

**Response:**
```json
{
  "agent_id": "BANKING",
  "agent_type": "BANKING",
  "status": "APPROVED",
  "result": {
    "status": "APPROVED",
    "transaction_id": "TXN_abc123",
    "amount": 50000
  },
  "risk_score": 0.1,
  "explanation": "Fund transfer processed successfully",
  "confidence": 0.95,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Health Check

**GET** `/health`

Returns agent health status.

## Configuration

### Environment Variables

- **AGENT_TYPE**: Type of agent (BANKING, FRAUD, GUARDRAIL, CLEARANCE, SCORING)
- **SERVER_PORT**: Port to run the agent on
- **AGENT_ENDPOINT**: Public endpoint URL for the agent
- **MCP_SERVER_URL**: URL of MCP Server (Layer 1)
- **AGENT_AUTO_REGISTER**: Whether to auto-register with MCP Server

## Integration with MCP Server

Agents automatically register with the MCP Server on startup (if `AGENT_AUTO_REGISTER=true`). The MCP Server can then route tasks to these agents based on agent type and capabilities.

## Docker Deployment

Each agent can be containerized and deployed independently:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o agent cmd/server/main.go

FROM alpine:latest
COPY --from=builder /app/agent .
CMD ["./agent"]
```

## Next Steps

This is **Layer 3: Agent Mesh**. The next layers to build:

- **Layer 4**: ML Models (Fraud detection models, scoring models)
- **Layer 5**: Banking Integrations (MB, NB, DWH connections)

## License

[Your License Here]

