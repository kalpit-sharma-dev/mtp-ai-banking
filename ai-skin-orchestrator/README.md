# Layer 2: AI Skin Orchestrator

The AI Skin Orchestrator is the intelligent layer that sits above the MCP Server (Layer 1) and provides:

- **Natural Language Understanding** - Parses user requests in natural language
- **Intent Recognition** - Extracts banking intents from user input
- **Context Enrichment** - Enriches requests with user history, behavior patterns, and risk indicators
- **Multi-Agent Coordination** - Coordinates multiple agents when needed
- **Response Merging** - Merges responses from multiple agents with conflict resolution
- **LLM Integration** - Uses Large Language Models for advanced understanding

## Architecture

```
User Request (Natural Language)
        |
        v
AI Skin Orchestrator (Layer 2)
  ├── Intent Parser (LLM/Rule-based)
  ├── Context Enricher
  │   ├── History Service
  │   ├── Behavior Analyzer
  │   └── Risk Calculator
  ├── Multi-Agent Coordinator
  └── Response Merger
        |
        v
MCP Server (Layer 1)
        |
        v
Agent Mesh
```

## Features

✅ **Natural Language Processing** - Understands user requests in plain English  
✅ **Intent Extraction** - Identifies banking operations from text  
✅ **Context Enrichment** - Adds user profile, transaction history, behavior patterns  
✅ **Risk Assessment** - Calculates fraud, credit, velocity, and amount risks  
✅ **LLM Integration** - Optional OpenAI/Anthropic integration for advanced parsing  
✅ **Multi-Agent Support** - Coordinates multiple agents for complex requests  
✅ **Conflict Resolution** - Resolves conflicts between agent responses  
✅ **MCP Client** - Communicates with Layer 1 (MCP Server)  

## Prerequisites

- Go 1.21 or higher
- Layer 1 (MCP Server) running on port 8080
- (Optional) OpenAI API key for LLM features

## Installation

1. Navigate to the orchestrator directory:
```bash
cd ai-skin-orchestrator
```

2. Install dependencies:
```bash
go mod download
```

3. Copy environment file:
```bash
cp .env.example .env
```

4. Update `.env` with your configuration:
   - Set `MCP_SERVER_URL` to point to Layer 1
   - (Optional) Set `LLM_API_KEY` for LLM features
   - Adjust other settings as needed

5. Run the orchestrator:
```bash
go run cmd/server/main.go
```

The orchestrator will start on `http://localhost:8081`

## API Endpoints

### Process Request

**POST** `/api/v1/process`

Processes a user request through the full orchestration pipeline.

**Request:**
```json
{
  "user_id": "U10001",
  "channel": "MB",
  "input": "Transfer 50000 rupees to account XXXX4321 via NEFT",
  "input_type": "natural_language",
  "session_id": "sess_abc123"
}
```

**Response:**
```json
{
  "status": "APPROVED",
  "final_result": {
    "status": "APPROVED",
    "message": "Transaction processed successfully"
  },
  "risk_score": 0.12,
  "explanation": "Request evaluated by 1 agents. All agents agree on the decision.",
  "agent_responses": [
    {
      "agent_id": "mcp-agent",
      "agent_type": "ORCHESTRATED",
      "status": "APPROVED",
      "risk_score": 0.12,
      "explidence": "Transaction is within user limits and behavior pattern is normal."
    }
  ]
}
```

### Health Check

**GET** `/health`

Returns service health status.

## Example Usage

### Natural Language Request

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

### Structured Request

```bash
curl -X POST http://localhost:8081/api/v1/process \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "user_id": "U10001",
    "channel": "MB",
    "input": "{\"intent\": \"TRANSFER_NEFT\", \"entities\": {\"amount\": 50000, \"to_account\": \"XXXX4321\"}}",
    "input_type": "structured"
  }'
```

## Configuration

### LLM Configuration

To enable LLM-based intent parsing:

1. Set `LLM_ENABLED=true` in `.env`
2. Set `LLM_API_KEY` with your OpenAI API key
3. Optionally adjust `LLM_MODEL`, `LLM_TEMPERATURE`, etc.

If LLM is disabled, the orchestrator falls back to rule-based parsing.

### MCP Server Connection

Ensure `MCP_SERVER_URL` points to your Layer 1 MCP Server:
```
MCP_SERVER_URL=http://localhost:8080
```

## How It Works

1. **User Request** → User sends natural language or structured input
2. **Intent Parsing** → Extracts intent and entities (amount, account, etc.)
3. **Context Enrichment** → Adds user profile, history, behavior patterns, risk indicators
4. **MCP Communication** → Sends enriched request to MCP Server (Layer 1)
5. **Agent Execution** → MCP Server routes to appropriate agent(s)
6. **Response Merging** → Merges agent responses (if multiple agents)
7. **Final Response** → Returns merged response to user

## Integration with Layer 1

The AI Skin Orchestrator communicates with the MCP Server via HTTP:

- Submits tasks via `POST /api/v1/submit-task`
- Retrieves results via `GET /api/v1/get-result/{taskID}`

The orchestrator enriches the context before sending to MCP, enabling more intelligent routing.

## Next Steps

This is **Layer 2: AI Skin Orchestrator**. The next layers to build:

- **Layer 3**: Agent Mesh (Individual agent implementations)
- **Layer 4**: ML Models (Fraud detection, scoring models)
- **Layer 5**: Banking Integrations (MB, NB, DWH connections)

## License

[Your License Here]

