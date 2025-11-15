#!/bin/bash

# Script to test each service individually

echo "=========================================="
echo "Testing All Services Individually"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Create logs directory
mkdir -p logs

# Function to test a service
test_service() {
    local name=$1
    local dir=$2
    local port=$3
    local env_vars=$4
    
    echo -e "${YELLOW}Testing $name...${NC}"
    
    # Build first
    cd "$dir" || return 1
    if [ -n "$env_vars" ]; then
        eval "$env_vars"
    fi
    
    if ! go build ./cmd/server 2>&1 | tee "../logs/${name}-build.log" | grep -q "error"; then
        echo -e "  ${GREEN}✓ Build successful${NC}"
    else
        echo -e "  ${RED}✗ Build failed${NC}"
        cd ..
        return 1
    fi
    
    # Start service
    if [ -n "$env_vars" ]; then
        eval "$env_vars go run cmd/server/main.go" > "../logs/${name}.log" 2>&1 &
    else
        go run cmd/server/main.go > "../logs/${name}.log" 2>&1 &
    fi
    local pid=$!
    echo $pid > "../logs/${name}.pid"
    
    # Wait and check health
    sleep 4
    if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓ Service running on port $port${NC}"
        cd ..
        return 0
    else
        echo -e "  ${RED}✗ Service failed to start or not responding${NC}"
        echo -e "  ${YELLOW}Check logs/${name}.log for details${NC}"
        cd ..
        return 1
    fi
}

# Test each service
echo "1. Layer 1: MCP Server"
test_service "mcp-server" "mcp-server" "8080" ""
echo ""

echo "2. Layer 2: AI Skin Orchestrator"
test_service "orchestrator" "ai-skin-orchestrator" "8081" ""
echo ""

echo "3. Layer 3: Banking Agent"
test_service "banking-agent" "agent-mesh" "8001" "export AGENT_TYPE=BANKING SERVER_PORT=8001 AGENT_ENDPOINT=http://localhost:8001"
echo ""

echo "4. Layer 5: Banking Integrations"
test_service "banking-integrations" "banking-integrations" "7000" ""
echo ""

echo "=========================================="
echo "Final Status"
echo "=========================================="
echo ""

for port in 8080:8081:8001:7000; do
    p=$(echo $port | cut -d: -f1)
    name=$(echo $port | cut -d: -f2)
    echo -n "$name (Port $p): "
    if curl -s "http://localhost:$p/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ RUNNING${NC}"
    else
        echo -e "${RED}✗ NOT RUNNING${NC}"
    fi
done

echo ""
echo "Check logs/ directory for detailed logs"
echo "To stop all services: pkill -f 'go run cmd/server/main.go'"

