# AI Banking Platform - Project Structure

## Complete Project Layout

```
C:/ZMTP/
├── mcp-server/                    # Layer 1: MCP Server (Port 8080)
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── controller/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── router/
│   │   ├── service/
│   │   └── utils/
│   ├── examples/
│   ├── go.mod
│   ├── README.md
│   └── LAYER_1_SUMMARY.md
│
├── ai-skin-orchestrator/          # Layer 2: AI Skin Orchestrator (Port 8081)
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── controller/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── router/
│   │   ├── service/
│   │   └── utils/
│   ├── go.mod
│   ├── README.md
│   └── LAYER_2_SUMMARY.md
│
├── agent-mesh/                    # Layer 3: Agent Mesh (Ports 8001-8005)
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── controller/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── router/
│   │   ├── service/
│   │   │   ├── agent_base.go
│   │   │   ├── banking_agent.go
│   │   │   ├── fraud_agent.go
│   │   │   ├── guardrail_agent.go
│   │   │   ├── clearance_agent.go
│   │   │   └── scoring_agent.go
│   │   └── utils/
│   ├── go.mod
│   ├── README.md
│   └── LAYER_3_SUMMARY.md
│
├── ml-models/                     # Layer 4: ML Models (Port 9000)
│   ├── app/
│   │   ├── main.py
│   │   ├── config.py
│   │   ├── models/
│   │   │   ├── fraud_model.py
│   │   │   ├── credit_model.py
│   │   │   └── risk_model.py
│   │   └── routers/
│   ├── models/ (trained models)
│   ├── train_models.py
│   ├── requirements.txt
│   ├── README.md
│   └── LAYER_4_SUMMARY.md
│
├── banking-integrations/          # Layer 5: Banking Integrations (Port 7000)
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── controller/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── router/
│   │   ├── service/
│   │   │   ├── mb_service.go
│   │   │   ├── nb_service.go
│   │   │   ├── dwh_service.go
│   │   │   └── banking_gateway.go
│   │   └── utils/
│   ├── go.mod
│   ├── README.md
│   └── LAYER_5_SUMMARY.md
│
├── logs/                          # Log files from all services
│
└── Documentation/
    ├── SYSTEM_READY.md
    ├── TEST_GUIDE.md
    ├── COMPLETE_SYSTEM_TEST.md
    ├── START_ALL.md
    ├── PROJECT_STRUCTURE.md (this file)
    ├── start-all.bat
    ├── start-all-layers.sh
    ├── test-all-services.sh
    └── quick-test.sh
```

## Layer Organization

All 5 layers are now in separate folders:

1. **mcp-server/** - Central orchestration (Golang)
2. **ai-skin-orchestrator/** - Intelligence layer (Golang)
3. **agent-mesh/** - Agent execution (Golang)
4. **ml-models/** - Machine learning (Python)
5. **banking-integrations/** - Data access (Golang)

## Starting Services

Each layer can be started independently from its own folder:

```bash
# Layer 1
cd mcp-server && go run cmd/server/main.go

# Layer 2
cd ai-skin-orchestrator && go run cmd/server/main.go

# Layer 3 (each agent)
cd agent-mesh && AGENT_TYPE=BANKING go run cmd/server/main.go

# Layer 4
cd ml-models && python -m app.main

# Layer 5
cd banking-integrations && go run cmd/server/main.go
```

## Communication Flow

```
User Request
    ↓
ai-skin-orchestrator/ (Layer 2)
    ↓
mcp-server/ (Layer 1)
    ↓
agent-mesh/ (Layer 3)
    ├── May call ml-models/ (Layer 4)
    └── May call banking-integrations/ (Layer 5)
    ↓
Response
```

## Benefits of This Structure

✅ **Modular** - Each layer is independent  
✅ **Scalable** - Deploy layers separately  
✅ **Maintainable** - Clear separation of concerns  
✅ **Testable** - Test each layer independently  
✅ **Organized** - Easy to navigate and understand  

