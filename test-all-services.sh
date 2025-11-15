#!/bin/bash

# Comprehensive test script for all services

API_KEY="test-api-key"

echo "=========================================="
echo "AI Banking Platform - Service Tests"
echo "=========================================="
echo ""

# Test Layer 1: MCP Server
echo "=== Layer 1: MCP Server ==="
echo "1. Health Check..."
if curl -s http://localhost:8080/health | grep -q "healthy"; then
    echo "   ✓ Health check passed"
else
    echo "   ✗ Health check failed"
    exit 1
fi

echo "2. Submit Task..."
TASK_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/submit-task \
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

TASK_ID=$(echo "$TASK_RESPONSE" | grep -o '"task_id":"[^"]*' | cut -d'"' -f4)
if [ -n "$TASK_ID" ]; then
    echo "   ✓ Task submitted: $TASK_ID"
    sleep 3
    echo "3. Get Task Result..."
    RESULT=$(curl -s -X GET "http://localhost:8080/api/v1/get-result/$TASK_ID" \
      -H "X-API-Key: $API_KEY")
    STATUS=$(echo "$RESULT" | grep -o '"status":"[^"]*' | cut -d'"' -f4)
    echo "   ✓ Task status: $STATUS"
else
    echo "   ✗ Task submission failed"
fi
echo ""

# Test Layer 2: AI Skin Orchestrator
echo "=== Layer 2: AI Skin Orchestrator ==="
echo "1. Health Check..."
if curl -s http://localhost:8081/health | grep -q "healthy"; then
    echo "   ✓ Health check passed"
else
    echo "   ✗ Health check failed"
fi

echo "2. Process Natural Language Request..."
ORCH_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/process \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "user_id": "U10001",
    "channel": "MB",
    "input": "Transfer 50000 rupees to account XXXX4321 via NEFT",
    "input_type": "natural_language"
  }')
if echo "$ORCH_RESPONSE" | grep -q "status"; then
    echo "   ✓ Request processed"
else
    echo "   ✗ Request failed"
fi
echo ""

# Test Layer 3: Agent Mesh
echo "=== Layer 3: Agent Mesh ==="
for port in 8001 8002 8003; do
    agent_name=""
    case $port in
        8001) agent_name="Banking Agent" ;;
        8002) agent_name="Fraud Agent" ;;
        8003) agent_name="Guardrail Agent" ;;
    esac
    if curl -s "http://localhost:$port/health" | grep -q "healthy"; then
        echo "   ✓ $agent_name (port $port) - OK"
    else
        echo "   ✗ $agent_name (port $port) - NOT RUNNING"
    fi
done
echo ""

# Test Layer 4: ML Models
echo "=== Layer 4: ML Models ==="
if curl -s http://localhost:9000/health | grep -q "healthy"; then
    echo "   ✓ ML Models Service - OK"
    echo "   Testing Fraud Prediction..."
    FRAUD_RESPONSE=$(curl -s -X POST http://localhost:9000/api/v1/fraud/predict \
      -H "Content-Type: application/json" \
      -d '{
        "amount": 50000,
        "transaction_count_24h": 3.0,
        "beneficiary_age_days": 5.0
      }')
    if echo "$FRAUD_RESPONSE" | grep -q "fraud_score"; then
        echo "   ✓ Fraud prediction successful"
    fi
else
    echo "   ✗ ML Models Service - NOT RUNNING"
fi
echo ""

# Test Layer 5: Banking Integrations
echo "=== Layer 5: Banking Integrations ==="
if curl -s http://localhost:7000/health | grep -q "healthy"; then
    echo "   ✓ Banking Integrations - OK"
    echo "   Testing Balance Inquiry..."
    BALANCE_RESPONSE=$(curl -s -X POST http://localhost:7000/api/v1/balance \
      -H "Content-Type: application/json" \
      -H "X-API-Key: $API_KEY" \
      -d '{
        "user_id": "U10001",
        "account_id": "ACC_001",
        "channel": "MB"
      }')
    if echo "$BALANCE_RESPONSE" | grep -q "balance"; then
        echo "   ✓ Balance inquiry successful"
    fi
else
    echo "   ✗ Banking Integrations - NOT RUNNING"
fi
echo ""

echo "=========================================="
echo "Tests Complete!"
echo "=========================================="

