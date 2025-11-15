# How to Start and Test All Layers

## Quick Start (Manual)

### Option 1: Using Multiple Terminals

Open 7 terminals and run:

**Terminal 1 - Layer 1 (MCP Server):**
```bash
cd C:/ZMTP/mcp-server
go run cmd/server/main.go
```

**Terminal 2 - Layer 2 (AI Skin Orchestrator):**
```bash
cd C:/ZMTP/ai-skin-orchestrator
go run cmd/server/main.go
```

**Terminal 3 - Layer 3 (Banking Agent):**
```bash
cd C:/ZMTP/agent-mesh
export AGENT_TYPE=BANKING SERVER_PORT=8001 AGENT_ENDPOINT=http://localhost:8001
go run cmd/server/main.go
```

**Terminal 4 - Layer 3 (Fraud Agent):**
```bash
cd C:/ZMTP/agent-mesh
export AGENT_TYPE=FRAUD SERVER_PORT=8002 AGENT_ENDPOINT=http://localhost:8002
go run cmd/server/main.go
```

**Terminal 5 - Layer 3 (Guardrail Agent):**
```bash
cd C:/ZMTP/agent-mesh
export AGENT_TYPE=GUARDRAIL SERVER_PORT=8003 AGENT_ENDPOINT=http://localhost:8003
go run cmd/server/main.go
```

**Terminal 6 - Layer 4 (ML Models - Optional):**
```bash
cd C:/ZMTP/ml-models
python -m venv venv
venv\Scripts\activate  # Windows
pip install -r requirements.txt
python -m app.main
```

**Terminal 7 - Layer 5 (Banking Integrations):**
```bash
cd C:/ZMTP/banking-integrations
go run cmd/server/main.go
```

### Option 2: Using Scripts (Linux/Mac/Git Bash)

**Start all services:**
```bash
./start-all-layers.sh
```

**Test all services:**
```bash
./test-all-services.sh
```

**Quick health check:**
```bash
./quick-test.sh
```

**Stop all services:**
```bash
./stop-all-layers.sh
```

### Option 3: Using PowerShell (Windows)

```powershell
.\start-and-test.ps1
```

## Prerequisites

1. **Redis** must be running:
   ```bash
   docker run -d -p 6379:6379 --name redis redis:latest
   ```

2. **Go modules** downloaded:
   ```bash
   cd C:/ZMTP && go mod download
   cd ai-skin-orchestrator && go mod download
   cd agent-mesh && go mod download
   cd banking-integrations && go mod download
   ```

3. **Python dependencies** (for ML Models):
   ```bash
   cd ml-models
   python -m venv venv
   venv\Scripts\activate
   pip install -r requirements.txt
   ```

## Testing Individual Services

### Test Layer 1 (MCP Server)
```bash
curl http://localhost:8080/health
curl -X POST http://localhost:8080/api/v1/submit-task \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{"user_id":"U10001","channel":"MB","intent":"CHECK_BALANCE","data":{}}'
```

### Test Layer 2 (Orchestrator)
```bash
curl http://localhost:8081/health
curl -X POST http://localhost:8081/api/v1/process \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{"user_id":"U10001","channel":"MB","input":"Check my balance","input_type":"natural_language"}'
```

### Test Layer 3 (Agents)
```bash
curl http://localhost:8001/health  # Banking Agent
curl http://localhost:8002/health  # Fraud Agent
curl http://localhost:8003/health  # Guardrail Agent
```

### Test Layer 4 (ML Models)
```bash
curl http://localhost:9000/health
curl -X POST http://localhost:9000/api/v1/fraud/predict \
  -H "Content-Type: application/json" \
  -d '{"amount":50000,"transaction_count_24h":3.0}'
```

### Test Layer 5 (Banking Integrations)
```bash
curl http://localhost:7000/health
curl -X POST http://localhost:7000/api/v1/balance \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{"user_id":"U10001","account_id":"ACC_001","channel":"MB"}'
```

## Expected Ports

| Service | Port | Status |
|---------|------|--------|
| MCP Server | 8080 | Required |
| AI Skin Orchestrator | 8081 | Required |
| Banking Agent | 8001 | Required |
| Fraud Agent | 8002 | Optional |
| Guardrail Agent | 8003 | Optional |
| ML Models | 9000 | Optional |
| Banking Integrations | 7000 | Optional |

## Troubleshooting

1. **Port conflicts**: Change ports in `.env` files
2. **Redis errors**: Start Redis or services will use in-memory storage
3. **Module errors**: Run `go mod download` in each directory
4. **Python errors**: Ensure virtual environment is activated

## Next Steps

Once all services are running:
1. Test end-to-end flow
2. Check logs in `logs/` directory
3. Monitor service health
4. Test with real banking scenarios

