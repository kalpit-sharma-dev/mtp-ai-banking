#!/bin/bash

# Test script for MCP Server API
# Make sure the server is running on localhost:8080

BASE_URL="http://localhost:8080/api/v1"
API_KEY="test-api-key"

echo "=== Testing MCP Server API ==="
echo ""

# 1. Health Check
echo "1. Health Check"
curl -s http://localhost:8080/health | jq .
echo ""
echo ""

# 2. Create Session
echo "2. Creating Session"
SESSION_RESPONSE=$(curl -s -X POST "$BASE_URL/create-session" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "user_id": "U10001",
    "channel": "MB"
  }')
echo "$SESSION_RESPONSE" | jq .
SESSION_ID=$(echo "$SESSION_RESPONSE" | jq -r '.session_id')
echo "Session ID: $SESSION_ID"
echo ""
echo ""

# 3. Submit Task - NEFT Transfer
echo "3. Submitting NEFT Transfer Task"
TASK_RESPONSE=$(curl -s -X POST "$BASE_URL/submit-task" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "{
    \"session_id\": \"$SESSION_ID\",
    \"user_id\": \"U10001\",
    \"channel\": \"MB\",
    \"intent\": \"TRANSFER_NEFT\",
    \"data\": {
      \"amount\": 50000,
      \"to_account\": \"XXXX4321\",
      \"ifsc\": \"BANK0001234\"
    }
  }")
echo "$TASK_RESPONSE" | jq .
TASK_ID=$(echo "$TASK_RESPONSE" | jq -r '.task_id')
echo "Task ID: $TASK_ID"
echo ""
echo ""

# 4. Wait and Get Task Result
echo "4. Waiting 2 seconds, then getting task result..."
sleep 2
curl -s -X GET "$BASE_URL/get-result/$TASK_ID" \
  -H "X-API-Key: $API_KEY" | jq .
echo ""
echo ""

# 5. Submit Task - Check Balance
echo "5. Submitting Check Balance Task"
TASK_RESPONSE2=$(curl -s -X POST "$BASE_URL/submit-task" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "{
    \"session_id\": \"$SESSION_ID\",
    \"user_id\": \"U10001\",
    \"channel\": \"MB\",
    \"intent\": \"CHECK_BALANCE\",
    \"data\": {}
  }")
echo "$TASK_RESPONSE2" | jq .
TASK_ID2=$(echo "$TASK_RESPONSE2" | jq -r '.task_id')
echo ""
echo ""

# 6. Get Task Result
echo "6. Getting Check Balance Result"
sleep 2
curl -s -X GET "$BASE_URL/get-result/$TASK_ID2" \
  -H "X-API-Key: $API_KEY" | jq .
echo ""
echo ""

# 7. Register Agent
echo "7. Registering a Custom Agent"
AGENT_RESPONSE=$(curl -s -X POST "$BASE_URL/register-agent" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "name": "Custom Payment Agent",
    "type": "PAYMENT",
    "endpoint": "http://localhost:8006",
    "capabilities": ["PAYMENT_PROCESSING", "WALLET_OPERATIONS"]
  }')
echo "$AGENT_RESPONSE" | jq .
echo ""
echo ""

# 8. Get All Agents
echo "8. Getting All Registered Agents"
curl -s -X GET "$BASE_URL/agents" \
  -H "X-API-Key: $API_KEY" | jq .
echo ""
echo ""

# 9. Get Session
echo "9. Getting Session Details"
curl -s -X GET "$BASE_URL/get-session/$SESSION_ID" \
  -H "X-API-Key: $API_KEY" | jq .
echo ""
echo ""

# 10. Get Rules
echo "10. Getting Routing Rules"
curl -s -X GET "$BASE_URL/rules" \
  -H "X-API-Key: $API_KEY" | jq .
echo ""
echo ""

echo "=== Test Complete ==="

