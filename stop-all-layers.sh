#!/bin/bash

# Script to stop all layers of the AI Banking Platform

echo "Stopping all AI Banking Platform services..."

# Read PIDs and kill processes
if [ -d "logs" ]; then
    for pidfile in logs/*.pid; do
        if [ -f "$pidfile" ]; then
            pid=$(cat "$pidfile")
            name=$(basename "$pidfile" .pid)
            if kill -0 "$pid" 2>/dev/null; then
                echo "Stopping $name (PID: $pid)..."
                kill "$pid" 2>/dev/null
            fi
        fi
    done
fi

# Kill any remaining Go/Python processes
pkill -f "go run cmd/server/main.go" 2>/dev/null
pkill -f "python -m app.main" 2>/dev/null
pkill -f "uvicorn app.main:app" 2>/dev/null

echo "All services stopped."

