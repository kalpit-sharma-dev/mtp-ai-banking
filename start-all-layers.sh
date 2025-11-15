#!/bin/bash

# Script to start all layers of the AI Banking Platform

echo "=========================================="
echo "Starting AI Banking Platform - All Layers"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Redis is running (required for Layer 1)
echo -e "${YELLOW}Checking Redis...${NC}"
if ! redis-cli ping > /dev/null 2>&1; then
    echo -e "${RED}Redis is not running. Starting Redis...${NC}"
    docker run -d -p 6379:6379 --name redis redis:latest 2>/dev/null || echo "Redis container may already exist"
    sleep 2
else
    echo -e "${GREEN}Redis is running${NC}"
fi

# Function to start a service
start_service() {
    local name=$1
    local dir=$2
    local port=$3
    local cmd=$4
    
    echo -e "${YELLOW}Starting $name on port $port...${NC}"
    cd "$dir" || exit 1
    
    # Start in background
    eval "$cmd" > "../logs/${name}.log" 2>&1 &
    local pid=$!
    echo $pid > "../logs/${name}.pid"
    
    # Wait for service to be ready
    echo -e "${YELLOW}Waiting for $name to be ready...${NC}"
    for i in {1..30}; do
        if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
            echo -e "${GREEN}$name is ready!${NC}"
            cd ..
            return 0
        fi
        sleep 1
    done
    
    echo -e "${RED}$name failed to start${NC}"
    cd ..
    return 1
}

# Create logs directory
mkdir -p logs

# Start Layer 1: MCP Server
start_service "MCP-Server" "mcp-server" "8080" "go run cmd/server/main.go"

# Start Layer 2: AI Skin Orchestrator
start_service "AI-Skin-Orchestrator" "ai-skin-orchestrator" "8081" "go run cmd/server/main.go"

# Start Layer 3: Agent Mesh (Banking Agent)
start_service "Banking-Agent" "agent-mesh" "8001" "AGENT_TYPE=BANKING SERVER_PORT=8001 AGENT_ENDPOINT=http://localhost:8001 go run cmd/server/main.go"

# Start Layer 3: Agent Mesh (Fraud Agent)
start_service "Fraud-Agent" "agent-mesh" "8002" "AGENT_TYPE=FRAUD SERVER_PORT=8002 AGENT_ENDPOINT=http://localhost:8002 go run cmd/server/main.go"

# Start Layer 3: Agent Mesh (Guardrail Agent)
start_service "Guardrail-Agent" "agent-mesh" "8003" "AGENT_TYPE=GUARDRAIL SERVER_PORT=8003 AGENT_ENDPOINT=http://localhost:8003 go run cmd/server/main.go"

# Start Layer 4: ML Models (Python - requires venv)
echo -e "${YELLOW}Starting ML Models Service...${NC}"
cd ml-models || exit 1
if [ -d "venv" ]; then
    source venv/bin/activate 2>/dev/null || source venv/Scripts/activate 2>/dev/null
    python -m app.main > "../logs/ML-Models.log" 2>&1 &
    echo $! > "../logs/ML-Models.pid"
    cd ..
    
    # Wait for ML service
    echo -e "${YELLOW}Waiting for ML Models to be ready...${NC}"
    for i in {1..30}; do
        if curl -s "http://localhost:9000/health" > /dev/null 2>&1; then
            echo -e "${GREEN}ML Models is ready!${NC}"
            break
        fi
        sleep 1
    done
else
    echo -e "${RED}ML Models venv not found. Skipping...${NC}"
    cd ..
fi

# Start Layer 5: Banking Integrations
start_service "Banking-Integrations" "banking-integrations" "7000" "go run cmd/server/main.go"

echo ""
echo -e "${GREEN}=========================================="
echo "All layers started!"
echo "==========================================${NC}"
echo ""
echo "Services running:"
echo "  - MCP Server: http://localhost:8080"
echo "  - AI Skin Orchestrator: http://localhost:8081"
echo "  - Banking Agent: http://localhost:8001"
echo "  - Fraud Agent: http://localhost:8002"
echo "  - Guardrail Agent: http://localhost:8003"
echo "  - ML Models: http://localhost:9000"
echo "  - Banking Integrations: http://localhost:7000"
echo ""
echo "Logs are in the 'logs/' directory"
echo "To stop all services, run: ./stop-all-layers.sh"
echo ""

