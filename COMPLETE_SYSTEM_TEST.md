# Complete System Test Guide

## 🚀 Quick Start Commands

### Windows (Command Prompt or PowerShell)

**Start all services:**
```cmd
start-all.bat
```

**Or manually in separate windows:**
```cmd
REM Window 1 - MCP Server
cd C:\ZMTP\mcp-server
go run cmd/server/main.go

REM Window 2 - AI Skin Orchestrator  
cd C:\ZMTP\ai-skin-orchestrator
go run cmd/server/main.go

REM Window 3 - Banking Agent
cd C:\ZMTP\agent-mesh
set AGENT_TYPE=BANKING
set SERVER_PORT=8001
set AGENT_ENDPOINT=http://localhost:8001
go run cmd/server/main.go

REM Window 4 - Fraud Agent
cd C:\ZMTP\agent-mesh
set AGENT_TYPE=FRAUD
set SERVER_PORT=8002
set AGENT_ENDPOINT=http://localhost:8002
go run cmd/server/main.go

REM Window 5 - Guardrail Agent
cd C:\ZMTP\agent-mesh
set AGENT_TYPE=GUARDRAIL
set SERVER_PORT=8003
set AGENT_ENDPOINT=http://localhost:8003
go run cmd/server/main.go

REM Window 6 - Banking Integrations
cd C:\ZMTP\banking-integrations
go run cmd/server/main.go
```

### Linux/Mac/Git Bash

**Start all services:**
```bash
./start-all-layers.sh
```

**Test all services:**
```bash
./test-all-services.sh
```

## 📋 Service Status Check

After starting services, wait 5-10 seconds, then run:

```bash
./quick-test.sh
```

Or manually check each service:

```bash
curl http://localhost:8080/health  # MCP Server
curl http://localhost:8081/health  # Orchestrator
curl http://localhost:8001/health  # Banking Agent
curl http://localhost:8002/health  # Fraud Agent
curl http://localhost:8003/health  # Guardrail Agent
curl http://localhost:7000/health  # Banking Integrations
curl http://localhost:9000/health  # ML Models (if started)
```

## 🧪 Integration Test Scenarios

### Scenario 1: Complete Transaction Flow

**Step 1: Submit task via MCP Server**
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
      "to_account": "XXXX4321",
      "ifsc": "BANK0001234"
    }
  }'
```

**Step 2: Get result (use task_id from Step 1)**
```bash
curl -X GET http://localhost:8080/api/v1/get-result/{task_id} \
  -H "X-API-Key: test-api-key"
```

### Scenario 2: Natural Language Processing

**Process natural language request:**
```bash
curl -X POST http://localhost:8081/api/v1/process \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "user_id": "U10001",
    "channel": "MB",
    "input": "I want to transfer 50000 rupees to account number XXXX4321 using NEFT",
    "input_type": "natural_language"
  }'
```

### Scenario 3: Direct Agent Call

**Call Banking Agent directly:**
```bash
curl -X POST http://localhost:8001/api/v1/process \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "agent_id": "BANKING",
    "task": "CHECK_BALANCE",
    "input_context": {
      "user_id": "U10001",
      "data": {}
    }
  }'
```

### Scenario 4: ML Model Prediction

**Get fraud prediction:**
```bash
curl -X POST http://localhost:9000/api/v1/fraud/predict \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 50000,
    "transaction_count_24h": 3.0,
    "beneficiary_age_days": 5.0,
    "device_risk": 0.2
  }'
```

### Scenario 5: Banking Integration

**Get account balance:**
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

## ✅ Expected Results

### Successful Flow

1. **Task Submission** → Returns `task_id` and `session_id`
2. **Task Processing** → Status changes from `PENDING` → `PROCESSING` → `COMPLETED`
3. **Result Retrieval** → Returns `status`, `result`, `risk_score`, `explanation`

### Sample Successful Response

```json
{
  "task_id": "task_abc123",
  "status": "COMPLETED",
  "result": {
    "status": "APPROVED",
    "transaction_id": "TXN_xyz789",
    "message": "Transaction processed successfully"
  },
  "risk_score": 0.12,
  "explanation": "Transaction is within user limits and behavior pattern is normal.",
  "completed_at": "2024-01-15T10:30:05Z"
}
```

## 🔍 Troubleshooting

### Service Not Starting

1. **Check if port is in use:**
   ```bash
   netstat -ano | findstr :8080  # Windows
   lsof -i :8080                 # Linux/Mac
   ```

2. **Check logs:**
   - Look in `logs/` directory
   - Check console output in service windows

3. **Redis Connection:**
   - MCP Server needs Redis for session/task storage
   - Start Redis: `docker run -d -p 6379:6379 redis:latest`
   - Or services will use in-memory storage (data lost on restart)

### Common Issues

1. **"Module not found"** → Run `go mod download` in each directory
2. **"Port already in use"** → Change port in `.env` or kill process
3. **"Connection refused"** → Service not started or wrong port
4. **"Redis connection failed"** → Start Redis or continue without persistence

## 📊 Service Dependencies

```
Layer 2 (Orchestrator) → Requires Layer 1 (MCP Server)
Layer 1 (MCP Server) → Requires Redis (optional)
Layer 3 (Agents) → Auto-register with Layer 1
Layer 4 (ML Models) → Independent (optional)
Layer 5 (Banking) → Independent (optional)
```

## 🎯 Test Checklist

- [ ] Layer 1 (MCP Server) health check passes
- [ ] Layer 2 (Orchestrator) health check passes
- [ ] Layer 3 (Agents) health checks pass
- [ ] Layer 5 (Banking) health check passes
- [ ] Can submit task via MCP Server
- [ ] Can get task result
- [ ] Can process natural language request
- [ ] Agents auto-register with MCP Server
- [ ] End-to-end flow works

## 📝 Notes

- **Redis**: Required for persistent storage, but services work without it (in-memory)
- **ML Models**: Optional - agents work with mock predictions if ML service not running
- **All Agents**: Not required - system works with just Banking Agent
- **Banking Integrations**: Optional - agents use mock data if not running

## 🎉 Success Criteria

All tests pass when:
1. All health checks return 200 OK
2. Task submission returns task_id
3. Task result shows COMPLETED status
4. Natural language processing works
5. Agents respond to requests

