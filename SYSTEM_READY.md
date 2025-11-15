# ✅ AI Banking Platform - System Ready!

All 5 layers have been successfully built and are ready for testing.

## 📁 Project Structure

```
C:/ZMTP/
├── mcp-server/ (Layer 1, Port 8080)
├── ai-skin-orchestrator/ (Layer 2, Port 8081)
├── agent-mesh/ (Layer 3, Ports 8001-8005)
├── ml-models/ (Layer 4, Port 9000)
└── banking-integrations/ (Layer 5, Port 7000)
```

## 🚀 How to Start All Services

### Method 1: Windows Batch Script (Easiest)

Double-click or run:
```cmd
start-all.bat
```

This will open separate command windows for each service.

### Method 2: Manual Start (Recommended for Testing)

Open **7 separate terminal windows** and run:

**Window 1 - MCP Server:**
```cmd
cd C:\ZMTP\mcp-server
go run cmd/server/main.go
```

**Window 2 - AI Skin Orchestrator:**
```cmd
cd C:\ZMTP\ai-skin-orchestrator
go run cmd/server/main.go
```

**Window 3 - Banking Agent:**
```cmd
cd C:\ZMTP\agent-mesh
set AGENT_TYPE=BANKING
set SERVER_PORT=8001
set AGENT_ENDPOINT=http://localhost:8001
go run cmd/server/main.go
```

**Window 4 - Fraud Agent:**
```cmd
cd C:\ZMTP\agent-mesh
set AGENT_TYPE=FRAUD
set SERVER_PORT=8002
set AGENT_ENDPOINT=http://localhost:8002
go run cmd/server/main.go
```

**Window 5 - Guardrail Agent:**
```cmd
cd C:\ZMTP\agent-mesh
set AGENT_TYPE=GUARDRAIL
set SERVER_PORT=8003
set AGENT_ENDPOINT=http://localhost:8003
go run cmd/server/main.go
```

**Window 6 - Banking Integrations:**
```cmd
cd C:\ZMTP\banking-integrations
go run cmd/server/main.go
```

**Window 7 - ML Models (Optional):**
```cmd
cd C:\ZMTP\ml-models
python -m venv venv
venv\Scripts\activate
pip install -r requirements.txt
python -m app.main
```

## 🧪 Quick Test

After starting services, wait 5-10 seconds, then test:

### Test 1: Health Checks
```bash
curl http://localhost:8080/health  # MCP Server
curl http://localhost:8081/health  # Orchestrator
curl http://localhost:8001/health  # Banking Agent
curl http://localhost:7000/health  # Banking Integrations
```

### Test 2: Submit Task
```bash
curl -X POST http://localhost:8080/api/v1/submit-task \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d "{\"user_id\":\"U10001\",\"channel\":\"MB\",\"intent\":\"TRANSFER_NEFT\",\"data\":{\"amount\":50000,\"to_account\":\"XXXX4321\"}}"
```

### Test 3: Natural Language Processing
```bash
curl -X POST http://localhost:8081/api/v1/process \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d "{\"user_id\":\"U10001\",\"channel\":\"MB\",\"input\":\"Transfer 50000 rupees to XXXX4321\",\"input_type\":\"natural_language\"}"
```

## ⚠️ Important Notes

1. **Redis**: MCP Server works without Redis (uses in-memory storage), but data is lost on restart. For persistence:
   ```bash
   docker run -d -p 6379:6379 --name redis redis:latest
   ```

2. **Port Conflicts**: If a port is in use, change it in the respective `.env` file

3. **Dependencies**: Make sure Go modules are downloaded:
   ```bash
   cd C:\ZMTP && go mod download
   cd ai-skin-orchestrator && go mod download
   cd agent-mesh && go mod download
   cd banking-integrations && go mod download
   ```

4. **Python (ML Models)**: Optional - install if you want to test ML models:
   ```bash
   cd ml-models
   python -m venv venv
   venv\Scripts\activate
   pip install -r requirements.txt
   ```

## 📚 Documentation

- `TEST_GUIDE.md` - Detailed testing guide
- `COMPLETE_SYSTEM_TEST.md` - Complete test scenarios
- `START_ALL.md` - Starting instructions
- Each layer has its own `README.md` and `LAYER_X_SUMMARY.md`

## ✅ System Status

All 5 layers are **100% complete** and ready for testing:

- ✅ Layer 1: MCP Server
- ✅ Layer 2: AI Skin Orchestrator  
- ✅ Layer 3: Agent Mesh (5 agents)
- ✅ Layer 4: ML Models (3 models)
- ✅ Layer 5: Banking Integrations

## 🎯 Next Steps

1. Start all services using `start-all.bat` or manually
2. Wait 5-10 seconds for services to initialize
3. Run `./quick-test.sh` or test manually
4. Test end-to-end flow
5. Check logs in `logs/` directory if issues occur

## 🆘 Troubleshooting

See `TEST_GUIDE.md` and `COMPLETE_SYSTEM_TEST.md` for detailed troubleshooting steps.

---

**🎉 Your AI Banking Platform is ready!**

