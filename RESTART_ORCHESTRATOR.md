# Restart AI Skin Orchestrator Service

## Issue
The AI Assistant is showing "Network error" because the orchestrator service needs to be restarted to pick up the improvements.

## Solution

### Option 1: Restart All Services (Recommended)
1. Close all service windows (MCP Server, AI Skin Orchestrator, etc.)
2. Run `start-all.bat` again to restart all services

### Option 2: Restart Only Orchestrator
1. Find the "AI Skin Orchestrator" window
2. Press `Ctrl+C` to stop it
3. Run this command in a new terminal:
   ```bash
   cd C:\ZMTP\ai-skin-orchestrator
   go run cmd/server/main.go
   ```

## What Was Fixed
1. **Better Error Handling**: Unknown intents now return helpful error messages instead of crashing
2. **Improved Intent Recognition**: Added more keywords like "transfer", "send money", "how much", etc.
3. **Graceful Degradation**: Service now handles unrecognized inputs gracefully

## After Restart
1. Wait 5-10 seconds for the service to start
2. Check the console for: "Starting AI Skin Orchestrator (Layer 2)"
3. Test the AI Assistant again in the web UI

## Verify Service is Running
Run this command to check:
```bash
curl http://localhost:8081/health
```

You should see: `{"status":"healthy","service":"ai-skin-orchestrator"}`

