# Port Conflict Resolution Guide

## Problem
Error: `listen tcp 0.0.0.0:8001: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.`

This means a service is already running on that port.

## Quick Fix

### Option 1: Stop All Services (Recommended)
```bash
# Run the stop script
stop-all-services.bat
```

Or manually:
```bash
# Check what's using the port
netstat -ano | findstr :8001

# Kill the process (replace PID with actual process ID)
taskkill /F /PID <PID>
```

### Option 2: Check Ports Before Starting
```bash
# Check which ports are in use
check-ports.bat
```

### Option 3: Use Different Ports
If you want to run multiple instances, change the port in the environment variable:
```bash
set SERVER_PORT=8004
go run cmd/server/main.go
```

## Ports Used by AI Banking Platform

| Service | Port | Description |
|---------|------|-------------|
| MCP Server | 8080 | Layer 1 |
| AI Skin Orchestrator | 8081 | Layer 2 |
| Banking Agent | 8001 | Layer 3 |
| Fraud Agent | 8002 | Layer 3 |
| Guardrail Agent | 8003 | Layer 3 |
| Banking Integrations | 7000 | Layer 5 |
| ML Models | 9000 | Layer 4 |

## Best Practice

Always stop all services before starting new ones:
1. Run `stop-all-services.bat`
2. Wait 2-3 seconds
3. Run `start-all.bat`

## Troubleshooting

If ports are still in use after stopping:
1. Check Task Manager for `go.exe` processes
2. Close any command windows running services
3. Restart your terminal/command prompt
4. Run `stop-all-services.bat` again

