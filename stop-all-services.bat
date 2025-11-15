@echo off
REM Script to stop all AI Banking Platform services

echo ==========================================
echo Stopping All AI Banking Platform Services
echo ==========================================
echo.

echo Stopping services on ports 8080, 8081, 8001, 8002, 8003, 7000, 9000...

REM Kill processes on specific ports
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8080') do (
    echo Killing process on port 8080 (PID: %%a)
    taskkill /F /PID %%a >nul 2>&1
)

for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8081') do (
    echo Killing process on port 8081 (PID: %%a)
    taskkill /F /PID %%a >nul 2>&1
)

for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8001') do (
    echo Killing process on port 8001 (PID: %%a)
    taskkill /F /PID %%a >nul 2>&1
)

for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8002') do (
    echo Killing process on port 8002 (PID: %%a)
    taskkill /F /PID %%a >nul 2>&1
)

for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8003') do (
    echo Killing process on port 8003 (PID: %%a)
    taskkill /F /PID %%a >nul 2>&1
)

for /f "tokens=5" %%a in ('netstat -ano ^| findstr :7000') do (
    echo Killing process on port 7000 (PID: %%a)
    taskkill /F /PID %%a >nul 2>&1
)

for /f "tokens=5" %%a in ('netstat -ano ^| findstr :9000') do (
    echo Killing process on port 9000 (PID: %%a)
    taskkill /F /PID %%a >nul 2>&1
)

REM Also kill Go processes
echo.
echo Killing any remaining Go processes...
taskkill /F /IM go.exe >nul 2>&1
taskkill /F /FI "WINDOWTITLE eq Banking Agent*" >nul 2>&1
taskkill /F /FI "WINDOWTITLE eq Fraud Agent*" >nul 2>&1
taskkill /F /FI "WINDOWTITLE eq Guardrail Agent*" >nul 2>&1
taskkill /F /FI "WINDOWTITLE eq MCP Server*" >nul 2>&1
taskkill /F /FI "WINDOWTITLE eq AI Skin Orchestrator*" >nul 2>&1
taskkill /F /FI "WINDOWTITLE eq Banking Integrations*" >nul 2>&1

echo.
echo ==========================================
echo All services stopped!
echo ==========================================
echo.
pause

