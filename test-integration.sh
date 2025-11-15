#!/bin/bash

# Integration test script for AI Banking Platform

echo "=========================================="
echo "AI Banking Platform - Integration Tests"
echo "=========================================="
echo ""

API_KEY="test-api-key"
BASE_URL_MCP="http://localhost:8080"
BASE_URL_ORCHESTRATOR="http://localhost:8081"
BASE_URL_BANKING="http://localhost:7000"
BASE_URL_ML="http://localhost:9000"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test function
test_endpoint() {
    local name=$1
    local method=$2
    local url=$3
    local data=$4
    
    echo -e "${YELLOW}Testing: $name${NC}"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" -H "X-API-Key: $API_KEY" "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" -H "Content-Type: application/json" -H "X-API-Key: $API_KEY" -d "$data" "$url")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✓ $name - Success (HTTP $http_code)${NC}"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        echo -e "${RED}✗ $name - Failed (HTTP $http_code)${NC}"
        echo "$body"
        return 1
    fi
    echo ""
}

# Test Layer 1: MCP Server
echo "=== Testing Layer 1: MCP Server ==="
test_endpoint "MCP Health Check" "GET" "$BASE_URL_MCP/health" ""
test_endpoint "Register Agent" "POST" "$BASE_URL_MCP/api/v1/register-agent" '{
  "name": "Test Banking Agent",
  "type": "BANKING",
  "endpoint": "http://localhost:8001",
  "capabilities": ["CHECK_BALANCE", "FUND_TRANSFER"]
}'
echo ""

# Test Layer 2: AI Skin Orchestrator
echo "=== Testing Layer 2: AI Skin Orchestrator ==="
test_endpoint "Orchestrator Health Check" "GET" "$BASE_URL_ORCHESTRATOR/health" ""
test_endpoint "Process Natural Language Request" "POST" "$BASE_URL_ORCHESTRATOR/api/v1/process" '{
  "user_id": "U10001",
  "channel": "MB",
  "input": "Transfer 50000 rupees to account XXXX4321 via NEFT",
  "input_type": "natural_language"
}'
echo ""

# Test Layer 3: Agent Mesh
echo "=== Testing Layer 3: Agent Mesh ==="
test_endpoint "Banking Agent Health" "GET" "http://localhost:8001/health" ""
test_endpoint "Fraud Agent Health" "GET" "http://localhost:8002/health" ""
test_endpoint "Guardrail Agent Health" "GET" "http://localhost:8003/health" ""
echo ""

# Test Layer 4: ML Models
echo "=== Testing Layer 4: ML Models ==="
test_endpoint "ML Models Health" "GET" "$BASE_URL_ML/health" ""
test_endpoint "Fraud Prediction" "POST" "$BASE_URL_ML/api/v1/fraud/predict" '{
  "amount": 50000,
  "transaction_count_24h": 3.0,
  "beneficiary_age_days": 5.0,
  "device_risk": 0.2
}'
test_endpoint "Credit Scoring" "POST" "$BASE_URL_ML/api/v1/scoring/credit" '{
  "account_age_days": 365,
  "monthly_income": 50000,
  "total_balance": 150000,
  "delinquency_count": 0
}'
echo ""

# Test Layer 5: Banking Integrations
echo "=== Testing Layer 5: Banking Integrations ==="
test_endpoint "Banking Integrations Health" "GET" "$BASE_URL_BANKING/health" ""
test_endpoint "Get Balance (MB)" "POST" "$BASE_URL_BANKING/api/v1/balance" '{
  "user_id": "U10001",
  "account_id": "ACC_001",
  "channel": "MB"
}'
test_endpoint "Get Transaction History" "GET" "$BASE_URL_BANKING/api/v1/dwh/history/U10001?days=90" ""
echo ""

# Test End-to-End Flow
echo "=== Testing End-to-End Flow ==="
echo -e "${YELLOW}Testing complete flow: User Request → Orchestrator → MCP → Agent → Response${NC}"

# Step 1: Submit task via MCP
echo "Step 1: Submitting task to MCP Server..."
TASK_RESPONSE=$(curl -s -X POST "$BASE_URL_MCP/api/v1/submit-task" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "user_id": "U10001",
    "channel": "MB",
    "intent": "TRANSFER_NEFT",
    "data": {
      "amount": 50000,
      "to_account": "XXXX4321"
    }
  }')

TASK_ID=$(echo "$TASK_RESPONSE" | jq -r '.task_id' 2>/dev/null)
if [ -n "$TASK_ID" ] && [ "$TASK_ID" != "null" ]; then
    echo -e "${GREEN}✓ Task submitted: $TASK_ID${NC}"
    
    # Step 2: Get result
    sleep 3
    echo "Step 2: Getting task result..."
    RESULT_RESPONSE=$(curl -s -X GET "$BASE_URL_MCP/api/v1/get-result/$TASK_ID" \
      -H "X-API-Key: $API_KEY")
    
    STATUS=$(echo "$RESULT_RESPONSE" | jq -r '.status' 2>/dev/null)
    if [ "$STATUS" != "null" ]; then
        echo -e "${GREEN}✓ Task completed with status: $STATUS${NC}"
        echo "$RESULT_RESPONSE" | jq . 2>/dev/null || echo "$RESULT_RESPONSE"
    else
        echo -e "${RED}✗ Failed to get result${NC}"
    fi
else
    echo -e "${RED}✗ Failed to submit task${NC}"
fi

echo ""
echo "=========================================="
echo "Integration tests completed!"
echo "=========================================="

