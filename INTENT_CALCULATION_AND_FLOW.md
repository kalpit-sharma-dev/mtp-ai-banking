# Intent Calculation & System Flow Documentation

## 📊 Part 1: Intent Calculation

### How Intent is Calculated

Intent calculation happens in **Layer 2: AI Skin Orchestrator** using the `IntentParser` service. There are **three methods** for intent parsing:

---

### Method 1: Structured Input (Direct JSON)

**When**: Input type is `"structured"`

**Algorithm**:
```go
1. Parse JSON input directly
2. Extract "intent" field → Intent Type
3. Extract "entities" field → Entities (amount, account, etc.)
4. Confidence = 1.0 (100% - fully structured)
```

**Example**:
```json
Input: {"intent": "TRANSFER_NEFT", "entities": {"amount": 50000, "to_account": "XXXX4321"}}
Result: Intent=TRANSFER_NEFT, Confidence=1.0, Entities={amount:50000, to_account:XXXX4321}
```

---

### Method 2: LLM-Based Parsing (Advanced)

**When**: `LLM_ENABLED=true` and LLM service is available

**Algorithm**:
```go
1. Create prompt with user input
2. Call LLM (GPT-3.5-turbo) with prompt
3. LLM returns JSON: {intent, confidence, entities}
4. Parse LLM response
5. If LLM fails → Fallback to rule-based
```

**Prompt Template**:
```
Analyze the following banking request and extract:
1. Intent type (one of: TRANSFER_NEFT, TRANSFER_RTGS, TRANSFER_IMPS, TRANSFER_UPI, 
   CHECK_BALANCE, GET_STATEMENT, ADD_BENEFICIARY, APPLY_LOAN, CREDIT_SCORE)
2. Entities (amount, account number, beneficiary, etc.)
3. Confidence score (0.0 to 1.0)

User request: "{userInput}"

Respond in JSON format:
{
  "intent": "INTENT_TYPE",
  "confidence": 0.95,
  "entities": {
    "amount": 50000,
    "to_account": "XXXX4321"
  }
}
```

**Confidence**: Provided by LLM (typically 0.85-0.95)

---

### Method 3: Rule-Based Parsing (Default)

**When**: LLM is disabled or unavailable

**Algorithm**:

#### Step 1: Entity Extraction (Regex Patterns)

```go
// Extract Amount
Pattern: (?i)(?:rs\.?|₹|rupees?)?\s*(\d+(?:,\d{3})*(?:\.\d{2})?)
Examples: "50000", "₹50,000", "50,000 rupees" → amount: 50000

// Extract Account Number
Pattern: (?i)(?:account|acc|ac)\s*(?:no|number|#)?\s*:?\s*([\dX]{4,})
Examples: "account XXXX4321", "acc no 123456" → to_account: XXXX4321

// Extract IFSC
Pattern: (?i)ifsc\s*:?\s*([A-Z]{4}0[A-Z0-9]{6})
Examples: "IFSC BANK0001234" → ifsc: BANK0001234
```

#### Step 2: Intent Classification (Keyword Matching)

**Algorithm**: Case-insensitive keyword matching with priority order

```go
Input: strings.ToLower(userInput)
Confidence: Default 0.7, adjusted based on match

Switch Statement (Priority Order):
```

| Intent | Keywords | Confidence |
|--------|----------|------------|
| **TRANSFER_NEFT** | "neft", "transfer neft", "send via neft", "transfer", "send money", "pay" | 0.9 |
| **TRANSFER_RTGS** | "rtgs", "transfer rtgs" | 0.9 |
| **TRANSFER_IMPS** | "imps", "transfer imps" | 0.9 |
| **TRANSFER_UPI** | "upi", "pay via upi", "scan qr" | 0.9 |
| **CHECK_BALANCE** | "balance", "check balance", "account balance", "how much", "what is my balance" | 0.95 |
| **GET_STATEMENT** | "statement", "mini statement", "transaction history", "transactions", "history" | 0.9 |
| **ADD_BENEFICIARY** | "add beneficiary", "add payee", "save beneficiary", "beneficiary" | 0.9 |
| **APPLY_LOAN** | "loan", "apply loan", "personal loan" | 0.85 |
| **CREDIT_SCORE** | "credit score", "cibil score", "credit rating" | 0.85 |
| **UNKNOWN** | No match | 0.3 |

**Matching Logic**:
```go
func containsAny(s string, keywords []string) bool {
    for _, keyword := range keywords {
        if strings.Contains(s, keyword) {
            return true
        }
    }
    return false
}
```

**Example Calculation**:
```
Input: "Check my balance"
↓
Step 1: Extract entities → None found
Step 2: Check keywords → Contains "balance" → CHECK_BALANCE
Step 3: Confidence → 0.95 (from table)
Result: Intent=CHECK_BALANCE, Confidence=0.95, Entities={}
```

```
Input: "Transfer 50000 rupees to account XXXX4321 via NEFT"
↓
Step 1: Extract entities:
  - Amount: "50000" (from regex)
  - Account: "XXXX4321" (from regex)
Step 2: Check keywords → Contains "neft" → TRANSFER_NEFT
Step 3: Confidence → 0.9 (from table)
Result: Intent=TRANSFER_NEFT, Confidence=0.9, Entities={amount:"50000", to_account:"XXXX4321"}
```

---

### Intent Calculation Summary

| Method | Confidence Range | Speed | Accuracy |
|--------|-----------------|-------|----------|
| **Structured** | 1.0 (100%) | Fastest | Perfect |
| **LLM-Based** | 0.85-0.95 | Medium | High |
| **Rule-Based** | 0.3-0.95 | Fastest | Medium |

**Fallback Chain**: LLM → Rule-Based → UNKNOWN

---

## 🔄 Part 2: Complete System Flow (UI to Backend)

### End-to-End Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    LAYER 0: Web UI (Port 3000)                   │
│  User types: "Check my balance"                                  │
└───────────────────────────┬─────────────────────────────────────┘
                             │
                             │ POST /api/process
                             │ {user_id, channel, input, input_type}
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│         LAYER 2: AI Skin Orchestrator (Port 8081)               │
│                                                                  │
│  Step 1: IntentParser.ParseIntent()                             │
│    ├─ Input: "Check my balance"                                 │
│    ├─ Method: Rule-based (LLM disabled)                         │
│    ├─ Extract: Keywords → "balance"                              │
│    └─ Result: Intent=CHECK_BALANCE, Confidence=0.95             │
│                                                                  │
│  Step 2: ContextEnricher.EnrichContext()                        │
│    ├─ Get user profile (account age, balance, etc.)              │
│    ├─ Get transaction history (last 90 days)                    │
│    ├─ Analyze behavior patterns                                 │
│    └─ Calculate risk indicators                                 │
│                                                                  │
│  Step 3: MCPClient.SubmitTask()                                 │
│    └─ POST http://localhost:8080/api/v1/submit-task            │
│       {user_id, channel, intent, enriched_context}               │
└───────────────────────────┬─────────────────────────────────────┘
                             │
                             │ HTTP POST
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│            LAYER 1: MCP Server (Port 8080)                      │
│                                                                  │
│  Step 1: TaskController.SubmitTask()                            │
│    ├─ Validate request                                          │
│    └─ Orchestrator.ProcessTask()                                │
│                                                                  │
│  Step 2: Session Management                                     │
│    ├─ Get or create session                                     │
│    └─ Create task with session_id                               │
│                                                                  │
│  Step 3: ContextRouter.RouteTask()                              │
│    ├─ Build enriched context                                    │
│    ├─ Apply routing rules                                        │
│    └─ Select agent: BANKING (for CHECK_BALANCE)                 │
│                                                                  │
│  Step 4: Execute Task (Async)                                   │
│    └─ go executeTask() → Call agent endpoint                   │
└───────────────────────────┬─────────────────────────────────────┘
                             │
                             │ HTTP POST
                             │ POST http://localhost:8001/api/v1/execute
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│         LAYER 3: Agent Mesh - Banking Agent (Port 8001)         │
│                                                                  │
│  Step 1: AgentController.Execute()                              │
│    ├─ Receive: {task, input_context, session_id}                │
│    └─ BankingAgent.Process()                                    │
│                                                                  │
│  Step 2: BankingAgent.Process()                                 │
│    ├─ Extract intent: CHECK_BALANCE                              │
│    ├─ Extract user_id, account_id from context                 │
│    └─ Call Banking Integrations API                             │
│                                                                  │
│  Step 3: Call Banking Integrations                              │
│    └─ POST http://localhost:7000/api/v1/balance                │
│       {user_id, account_id, channel}                             │
└───────────────────────────┬─────────────────────────────────────┘
                             │
                             │ HTTP POST
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│      LAYER 5: Banking Integrations (Port 7000)                  │
│                                                                  │
│  Step 1: BankingController.GetBalance()                         │
│    └─ BankingGateway.GetBalance()                                │
│                                                                  │
│  Step 2: Channel Routing                                        │
│    ├─ Channel = "MB" → MBService.GetBalance()                   │
│    └─ Returns: {account_id, balance, currency, ...}              │
└───────────────────────────┬─────────────────────────────────────┘
                             │
                             │ Response: {balance: 50000, currency: "INR"}
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│         LAYER 3: Banking Agent (Response)                       │
│                                                                  │
│  Step 1: Format Response                                        │
│    ├─ Result: {balance: 50000, currency: "INR", ...}            │
│    ├─ Risk Score: 0.1 (low risk for balance check)             │
│    └─ Explanation: "Balance retrieved successfully"             │
│                                                                  │
│  Step 2: Return to MCP Server                                   │
│    └─ HTTP 200: {status, result, risk_score, explanation}        │
└───────────────────────────┬─────────────────────────────────────┘
                             │
                             │ Response
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│            LAYER 1: MCP Server (Response)                       │
│                                                                  │
│  Step 1: Update Task Result                                     │
│    ├─ TaskManager.UpdateTaskResult()                            │
│    └─ Store: {status: "COMPLETED", result: {...}, ...}          │
│                                                                  │
│  Step 2: Return Task ID                                         │
│    └─ HTTP 202: {task_id, session_id, status}                   │
└───────────────────────────┬─────────────────────────────────────┘
                             │
                             │ Task ID
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│         LAYER 2: AI Skin Orchestrator (Polling)                 │
│                                                                  │
│  Step 1: Wait 2 seconds (for async processing)                 │
│                                                                  │
│  Step 2: MCPClient.GetTaskResult()                              │
│    └─ GET http://localhost:8080/api/v1/get-result/{taskID}       │
│                                                                  │
│  Step 3: Receive Agent Response                                 │
│    └─ {task_id, status, result, risk_score, explanation}         │
│                                                                  │
│  Step 4: ResponseMerger.MergeResponses()                        │
│    ├─ Convert to MergedResponse                                 │
│    └─ Format: {status, final_result, risk_score, explanation}   │
└───────────────────────────┬─────────────────────────────────────┘
                             │
                             │ HTTP 200
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    LAYER 0: Web UI (Response)                   │
│                                                                  │
│  Step 1: Receive Response                                       │
│    └─ {status: "COMPLETED", final_result: {balance: 50000}, ...}│
│                                                                  │
│  Step 2: Format Display                                         │
│    ├─ Extract balance from final_result                          │
│    ├─ Format: "Your account balance is ₹50,000.00"               │
│    └─ Display in chat UI                                         │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📋 Detailed Flow Steps

### Phase 1: UI → AI Skin Orchestrator

**File**: `web-ui/src/pages/AIAssistant.jsx`
```javascript
User Input: "Check my balance"
↓
orchestratorAPI.processRequest("Check my balance", "U10001", "MB")
↓
POST http://localhost:8081/api/v1/process
{
  user_id: "U10001",
  channel: "MB",
  input: "Check my balance",
  input_type: "natural_language"
}
```

---

### Phase 2: AI Skin Orchestrator - Intent Parsing

**File**: `ai-skin-orchestrator/internal/service/intent_parser.go`

```go
ParseIntent("Check my balance", "natural_language")
↓
parseWithRules("Check my balance")
↓
1. Extract entities: None
2. Check keywords: Contains "balance" → CHECK_BALANCE
3. Confidence: 0.95
↓
Result: Intent{Type: CHECK_BALANCE, Confidence: 0.95, Entities: {}}
```

---

### Phase 3: AI Skin Orchestrator - Context Enrichment

**File**: `ai-skin-orchestrator/internal/service/context_enricher.go`

```go
EnrichContext(userID, sessionID, channel, intent)
↓
1. Get user profile: {account_age: 365, balance: 50000, ...}
2. Get transaction history: Last 90 days
3. Analyze behavior: {avg_amount: 5000, peak_hours: [10, 14, 18], ...}
4. Calculate risks: {fraud_risk: 0.1, credit_risk: 0.2, ...}
↓
Result: EnrichedContext{...}
```

---

### Phase 4: AI Skin Orchestrator → MCP Server

**File**: `ai-skin-orchestrator/internal/service/mcp_client.go`

```go
SubmitTask(userRequest, intent, enrichedContext)
↓
POST http://localhost:8080/api/v1/submit-task
{
  user_id: "U10001",
  channel: "MB",
  intent: "CHECK_BALANCE",
  context: {...enriched_context...}
}
↓
Response: {task_id: "task_abc123", status: "PENDING"}
```

---

### Phase 5: MCP Server - Task Routing

**File**: `mcp-server/internal/service/context_router.go`

```go
RouteTask(task, session)
↓
1. Build enriched context
2. Apply routing rules
3. Intent = "CHECK_BALANCE" → routeByIntent()
4. Switch case: CHECK_BALANCE → AgentTypeBanking
5. Find agent: Banking Agent (port 8001)
↓
Result: RoutingDecision{SelectedAgentID: "banking_agent_1", ...}
```

---

### Phase 6: MCP Server → Agent (Banking/Fraud/Scoring)

**File**: `mcp-server/internal/service/orchestrator.go`

```go
executeTask(task, decision)
↓
1. Get agent from registry: Selected Agent (Banking/Fraud/Scoring)
2. Prepare request:
   {
     agent_id: "agent_id",
     task: "INTENT",
     input_context: {user_id, intent, data, ...}
   }
3. POST http://localhost:{agent_port}/api/v1/execute
```

---

### Phase 7A: Fraud Agent Processing (For Transfers)

**File**: `agent-mesh/internal/service/fraud_agent.go`

```go
Process(inputContext)
↓
1. Extract transaction details
2. calculateFraudScore()
   ├─ Try ML Models Service (if enabled)
   │  └─ POST http://localhost:9000/api/v1/fraud/predict  ✅ ML MODEL CALL
   │     └─ XGBoost Model → Returns fraud_score
   └─ Fallback to rule-based (if ML fails)
↓
Response: {fraud_score: 0.15, risk_level: "LOW", ...}
```

---

### Phase 7B: Scoring Agent Processing (For Credit/Risk)

**File**: `agent-mesh/internal/service/scoring_agent.go`

```go
Process(inputContext)
↓
1. Determine score type: CREDIT/FRAUD/RISK
2. calculateCreditScore() or calculateRiskScore()
   ├─ Try ML Models Service (if enabled)
   │  ├─ POST http://localhost:9000/api/v1/scoring/credit  ✅ ML MODEL CALL
   │  │  └─ Random Forest Model → Returns credit_score
   │  └─ POST http://localhost:9000/api/v1/scoring/risk  ✅ ML MODEL CALL
   │     └─ Ensemble Model → Returns overall_risk
   └─ Fallback to rule-based (if ML fails)
↓
Response: {credit_score: 750, risk_category: "GOOD", ...}
```

---

### Phase 7C: Banking Agent Processing (For Balance/Statement)

**File**: `agent-mesh/internal/service/banking_agent.go`

```go
Process(inputContext)
↓
1. Extract intent: CHECK_BALANCE
2. Extract user_id, account_id
3. Call Banking Integrations:
   POST http://localhost:7000/api/v1/balance
   {user_id: "U10001", account_id: "ACC_001", channel: "MB"}
↓
Response: {balance: 50000, currency: "INR", ...}
↓
Format response:
{
  status: "COMPLETED",
  result: {balance: 50000, currency: "INR"},
  risk_score: 0.1,
  explanation: "Balance retrieved successfully"
}
```

---

### Phase 8: Banking Agent → MCP Server

**File**: `mcp-server/internal/service/orchestrator.go`

```go
Receive agent response
↓
UpdateTaskResult(taskID, result, riskScore, explanation)
↓
Task stored with:
{
  task_id: "task_abc123",
  status: "COMPLETED",
  result: {balance: 50000, currency: "INR"},
  risk_score: 0.1,
  explanation: "..."
}
```

---

### Phase 9: MCP Server → AI Skin Orchestrator

**File**: `ai-skin-orchestrator/internal/service/mcp_client.go`

```go
GetTaskResult(taskID)
↓
Wait 2 seconds (for async processing)
↓
GET http://localhost:8080/api/v1/get-result/task_abc123
↓
Response: {
  task_id: "task_abc123",
  status: "COMPLETED",
  result: {balance: 50000, currency: "INR"},
  risk_score: 0.1,
  explanation: "..."
}
↓
Convert to AgentResponse
```

---

### Phase 10: AI Skin Orchestrator - Response Merging

**File**: `ai-skin-orchestrator/internal/service/response_merger.go`

```go
MergeResponses([agentResponse])
↓
singleResponseToMerged(agentResponse)
↓
Result: MergedResponse{
  status: "COMPLETED",
  final_result: {balance: 50000, currency: "INR"},
  risk_score: 0.1,
  explanation: "...",
  agent_responses: [...]
}
```

---

### Phase 11: AI Skin Orchestrator → UI

**File**: `web-ui/src/pages/AIAssistant.jsx`

```javascript
Receive response:
{
  status: "COMPLETED",
  final_result: {balance: 50000, currency: "INR"},
  risk_score: 0.1,
  explanation: "..."
}
↓
Extract balance: final_result.balance = 50000
↓
Format message: "Your account balance is ₹50,000.00"
↓
Display in chat UI
```

---

## 🔀 Agent Routing Logic

### Routing Decision Tree

```
Task Intent
    │
    ├─ TRANSFER_* (NEFT, RTGS, IMPS, UPI)
    │   ├─ High Risk? → Guardrail Agent
    │   ├─ Fraud Flag? → Fraud Agent
    │   │   └─ Calls ML Models: POST /api/v1/fraud/predict ✅
    │   └─ Default → Banking Agent
    │
    ├─ CHECK_BALANCE, GET_STATEMENT
    │   └─ → Banking Agent
    │
    ├─ ADD_BENEFICIARY
    │   └─ → Guardrail Agent
    │
    ├─ APPLY_LOAN
    │   └─ → Clearance Agent
    │
    └─ CREDIT_SCORE, RISK_ASSESSMENT
        └─ → Scoring Agent
            ├─ Calls ML Models: POST /api/v1/scoring/credit ✅
            └─ Calls ML Models: POST /api/v1/scoring/risk ✅
```

## 🤖 ML Models Integration

### When ML Models Are Called

1. **Fraud Agent** → Calls ML Models when processing transfers:
   - Endpoint: `POST http://localhost:9000/api/v1/fraud/predict`
   - Model: **XGBoost Classifier**
   - Returns: `fraud_score` (0.0-1.0)

2. **Scoring Agent** → Calls ML Models for scoring:
   - Credit: `POST http://localhost:9000/api/v1/scoring/credit`
     - Model: **Random Forest Regressor**
     - Returns: `credit_score` (300-850)
   - Risk: `POST http://localhost:9000/api/v1/scoring/risk`
     - Model: **Ensemble Model**
     - Returns: `overall_risk` (0.0-1.0)

### ML Models Flow

```
Transfer Request
    ↓
MCP Server → Routes to Fraud Agent
    ↓
Fraud Agent
    ├─ Extract features (amount, time, velocity, etc.)
    └─ POST http://localhost:9000/api/v1/fraud/predict  ✅
        ↓
    ML Models Service (Port 9000)
        ├─ Load XGBoost model
        ├─ Predict fraud probability
        └─ Return: {fraud_score: 0.15, risk_level: "LOW"}
        ↓
    Fraud Agent
        └─ Return fraud score to MCP Server
```

### Fallback Behavior

- If ML Models service is **disabled** → Use rule-based calculation
- If ML Models service is **unavailable** → Fallback to rule-based
- If ML Models service **errors** → Fallback to rule-based
- Rule-based calculations are always available as backup

### Routing Rules (Priority Order)

1. **Rule Engine** (if rules match)
2. **Intent-based routing** (default)
3. **Fallback to Banking Agent** (if no agent found)

---

## 📊 Data Flow Summary

| Layer | Input | Output | Protocol |
|-------|-------|--------|----------|
| **Web UI** | User text | HTTP POST | HTTP/REST |
| **AI Orchestrator** | User text | Intent + Context | HTTP/REST |
| **MCP Server** | Intent + Context | Task ID | HTTP/REST |
| **Agent Mesh** | Task + Context | Result + Risk | HTTP/REST |
| **Banking Integrations** | Request | Data | HTTP/REST |

---

## ⚡ Performance Characteristics

- **Intent Parsing**: < 10ms (rule-based), 200-500ms (LLM)
- **Context Enrichment**: < 50ms
- **Agent Routing**: < 20ms
- **Agent Execution**: 100-500ms (depends on backend)
- **Total End-to-End**: 500ms - 2s (typical)

---

## 🔐 Security Flow

1. **API Key Authentication**: All inter-service calls use `X-API-Key` header
2. **Session Management**: Tasks linked to sessions for tracking
3. **Risk Scoring**: Every transaction gets risk assessment
4. **Agent Validation**: Only registered agents can be called

---

This document provides a complete understanding of how intent is calculated and how data flows through the entire system from UI to backend and back.

