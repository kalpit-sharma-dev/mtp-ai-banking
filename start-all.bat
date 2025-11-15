@echo off
REM Batch script to start all layers on Windows

echo ==========================================
echo Starting AI Banking Platform - All Layers
echo ==========================================
echo.

REM Create logs directory
if not exist logs mkdir logs


REM Start Layer 2: AI Skin Orchestrator
echo Starting Layer 2: AI Skin Orchestrator (Port 8081)...
start "AI Skin Orchestrator" cmd /k "cd /d %~dp0ai-skin-orchestrator && go run cmd/server/main.go"

timeout /t 3 /nobreak >nul

REM Start Layer 1: MCP Server
echo Starting Layer 1: MCP Server (Port 8080)...
start "MCP Server" cmd /k "cd /d %~dp0mcp-server && go run cmd/server/main.go"

timeout /t 2 /nobreak >nul

REM Start Layer 3: Banking Agent
echo Starting Layer 3: Banking Agent (Port 8001)...
start "Banking Agent" cmd /k "cd /d %~dp0agent-mesh && set AGENT_TYPE=BANKING && set AGENT_NAME=Banking Agent && set SERVER_PORT=8001 && set AGENT_ENDPOINT=http://localhost:8001 && go run cmd/server/main.go"

timeout /t 2 /nobreak >nul

REM Start Layer 3: Fraud Agent
echo Starting Layer 3: Fraud Agent (Port 8002)...
start "Fraud Agent" cmd /k "cd /d %~dp0agent-mesh && set AGENT_TYPE=FRAUD && set AGENT_NAME=Fraud Agent && set SERVER_PORT=8002 && set AGENT_ENDPOINT=http://localhost:8002 && go run cmd/server/main.go"

timeout /t 2 /nobreak >nul

REM Start Layer 3: Guardrail Agent
echo Starting Layer 3: Guardrail Agent (Port 8003)...
start "Guardrail Agent" cmd /k "cd /d %~dp0agent-mesh && set AGENT_TYPE=GUARDRAIL && set AGENT_NAME=Guardrail Agent && set SERVER_PORT=8003 && set AGENT_ENDPOINT=http://localhost:8003 && go run cmd/server/main.go"

timeout /t 2 /nobreak >nul

REM Start Layer 4: ML Models
echo Starting Layer 4: ML Models Service (Port 9000)...
echo Installing Python dependencies (if needed)...
cd /d %~dp0ml-models
python -m pip install -q -r requirements.txt
if errorlevel 1 (
    echo WARNING: Failed to install dependencies. Please run: cd ml-models && pip install -r requirements.txt
)
cd /d %~dp0
start "ML Models Service" cmd /k "cd /d %~dp0ml-models && python -m app.main"

timeout /t 2 /nobreak >nul

REM Start Layer 5: Banking Integrations
echo Starting Layer 5: Banking Integrations (Port 7000)...
start "Banking Integrations" cmd /k "cd /d %~dp0banking-integrations && go run cmd/server/main.go"


echo.
echo ==========================================
echo All services started in separate windows!
echo ==========================================
echo.
echo Services:
echo   - MCP Server: http://localhost:8080
echo   - AI Skin Orchestrator: http://localhost:8081
echo   - Banking Agent: http://localhost:8001
echo   - Fraud Agent: http://localhost:8002
echo   - Guardrail Agent: http://localhost:8003
echo   - ML Models Service: http://localhost:9000
echo   - Banking Integrations: http://localhost:7000
echo.
echo Wait 5-10 seconds for services to start, then run: test-all-services.sh
echo.

pause

