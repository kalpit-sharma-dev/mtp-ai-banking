# Testing Guide - AI Banking Platform

This guide helps you test all 5 layers of the AI Banking Platform.

## Prerequisites

1. **Redis** (for Layer 1 - MCP Server)
   ```bash
   # Using Docker
   docker run -d -p 6379:6379 --name redis redis:latest
   
   # OR using local Redis
   redis-server
   ```

2. **Go 1.21+** installed
3. **Python 3.8+** (for Layer 4 - ML Models)

## Step-by-Step Testing

### Step 1: Start Layer 1 - MCP Server

**Terminal 1:**
```bash
cd C:/ZMTP/mcp-server
go run cmd/server/main.go
```

**Expected Output:**
```
{"level":"info","message":"Starting MCP Server for AI Banking Platform"}
{"level":"info","message":"Connected to Redis"}
{"level":"info","address":"0.0.0.0:8080","message":"MCP Server started"}
```

**Test:**
```bash
curl http://localhost:8080/health
# Should return: {"status":"healthy"}
```

### Step 2: Start Layer 2 - AI Skin Orchestrator

**Terminal 2:**
```bash
cd C:/ZMTP/ai-skin-orchestrator
go run cmd/server/main.go
```

**Expected Output:**
```
{"level":"info","message":"Starting AI Skin Orchestrator (Layer 2)"}
{"level":"info","address":"0.0.0.0:8081","message":"AI Skin Orchestrator started"}
```

**Test:**
```bash
curl http://localhost:8081/health
# Should return: {"status":"healthy","service":"ai-skin-orchestrator"}
```

### Step 3: Start Layer 3 - Agent Mesh

**Terminal 3 - Banking Agent:**
```bash
cd C:/ZMTP/agent-mesh
export AGENT_TYPE=BANKING
export SERVER_PORT=8001
export AGENT_ENDPOINT=http://localhost:8001
go run cmd/server/main.go
```

**Terminal 4 - Fraud Agent:**
```bash
cd C:/ZMTP/agent-mesh
export AGENT_TYPE=FRAUD
export SERVER_PORT=8002
export AGENT_ENDPOINT=http://localhost:8002
go run cmd/server/main.go
```

**Terminal 5 - Guardrail Agent:**
```bash
cd C:/ZMTP/agent-mesh
export AGENT_TYPE=GUARDRAIL
export SERVER_PORT=8003
export AGENT_ENDPOINT=http://localhost:8003
go run cmd/server/main.go
```

**Test:**
```bash
curl http://localhost:8001/health
curl http://localhost:8002/health
curl http://localhost:8003/health
```

### Step 4: Start Layer 4 - ML Models (Optional)

**Terminal 6:**
```bash
cd C:/ZMTP/ml-models
python -m venv venv
# Windows:
venv\Scripts\activate
# Linux/Mac:
source venv/bin/activate

pip install -r requirements.txt
python -m app.main
```

**Test:**
```bash
curl http://localhost:9000/health
```

### Step 5: Start Layer 5 - Banking Integrations

**Terminal 7:**
```bash
cd C:/ZMTP/banking-integrations
go run cmd/server/main.go
```

**Test:**
```bash
curl http://localhost:7000/health
```

## Integration Tests

### Test 1: Submit Task via MCP Server

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

**Expected Response:**
```json
{
  "task_id": "task_abc123",
  "session_id": "sess_xyz789",
  "status": "PENDING",
  "message": "Task submitted successfully"
}
```

### Test 2: Get Task Result

```bash
# Use task_id from previous response
curl -X GET http://localhost:8080/api/v1/get-result/task_abc123 \
  -H "X-API-Key: test-api-key"
```

**Expected Response:**
```json
{
  "task_id": "task_abc123",
  "status": "COMPLETED",
  "result": {
    "status": "APPROVED",
    "message": "Transaction processed successfully"
  },
  "risk_score": 0.12,
  "explanation": "Transaction is within user limits..."
}
```

### Test 3: Process Request via Orchestrator

```bash
curl -X POST http://localhost:8081/api/v1/process \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "user_id": "U10001",
    "channel": "MB",
    "input": "Transfer 50000 rupees to account XXXX4321 via NEFT",
    "input_type": "natural_language"
  }'
```

### Test 4: Get Balance from Banking Integrations

```bash
curl -X POST http://localhost:7000/api/v1/balance \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "user_id": "U10001",
    "account_id": "ACC_001",
    "channel": "MB"
  }'
```

### Test 5: Fraud Prediction from ML Models

```bash
curl -X POST http://localhost:9000/api/v1/fraud/predict \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 50000,
    "transaction_count_24h": 3.0,
    "beneficiary_age_days": 5.0
  }'
```

## End-to-End Test Flow

1. **User Request** → AI Skin Orchestrator (Layer 2)
2. **Orchestrator** → Enriches context → MCP Server (Layer 1)
3. **MCP Server** → Routes to appropriate Agent (Layer 3)
4. **Agent** → May call ML Models (Layer 4) or Banking Integrations (Layer 5)
5. **Response** → Flows back through layers to user

## Troubleshooting

### Redis Connection Error
- Ensure Redis is running: `redis-cli ping`
- Check Redis is on port 6379
- MCP Server will continue without Redis but won't persist data

### Port Already in Use
- Check which process is using the port
- Change port in `.env` file
- Kill the process: `lsof -ti:8080 | xargs kill` (Linux/Mac)

### Service Not Starting
- Check logs in `logs/` directory
- Verify Go modules are downloaded: `go mod download`
- Check Python dependencies: `pip install -r requirements.txt`

## Quick Test Script

Save this as `quick-test.sh`:

```bash
#!/bin/bash

echo "Testing all services..."

# Test each service
services=(
  "8080:MCP Server"
  "8081:AI Skin Orchestrator"
  "8001:Banking Agent"
  "8002:Fraud Agent"
  "8003:Guardrail Agent"
  "7000:Banking Integrations"
  "9000:ML Models"
)

for service in "${services[@]}"; do
  port="${service%%:*}"
  name="${service##*:}"
  
  if curl -s "http://localhost:$port/health" > /dev/null; then
    echo "✓ $name (port $port) - OK"
  else
    echo "✗ $name (port $port) - FAILED"
  fi
done
```

Run: `chmod +x quick-test.sh && ./quick-test.sh`

