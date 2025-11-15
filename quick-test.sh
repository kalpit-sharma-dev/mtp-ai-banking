#!/bin/bash

# Quick test script for all services

echo "=========================================="
echo "Quick Health Check - All Services"
echo "=========================================="
echo ""

# Test each service
services=(
  "8080:MCP Server"
  "8081:AI Skin Orchestrator"
  "8001:Banking Agent"
  "8002:Fraud Agent"
  "8003:Guardrail Agent"
  "7000:Banking Integrations"
  "9000:ML Models"
)

for service in "${services[@]}"; do
  port="${service%%:*}"
  name="${service##*:}"
  
  if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
    echo "✓ $name (port $port) - OK"
  else
    echo "✗ $name (port $port) - NOT RUNNING"
  fi
done

echo ""
echo "=========================================="

