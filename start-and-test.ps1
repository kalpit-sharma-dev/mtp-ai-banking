# PowerShell script to start and test all layers

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Starting AI Banking Platform - All Layers" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Create logs directory
New-Item -ItemType Directory -Force -Path "logs" | Out-Null

# Function to start a service
function Start-Service {
    param(
        [string]$Name,
        [string]$Directory,
        [int]$Port,
        [string]$Command
    )
    
    Write-Host "Starting $Name on port $Port..." -ForegroundColor Yellow
    
    Push-Location $Directory
    
    # Start process in background
    $process = Start-Process -FilePath "powershell" -ArgumentList "-Command", $Command -PassThru -WindowStyle Hidden
    
    # Save PID
    $process.Id | Out-File -FilePath "../logs/${Name}.pid" -Encoding ASCII
    
    # Wait for service to be ready
    Write-Host "Waiting for $Name to be ready..." -ForegroundColor Yellow
    $maxAttempts = 30
    $attempt = 0
    $ready = $false
    
    while ($attempt -lt $maxAttempts) {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:$Port/health" -Method GET -TimeoutSec 1 -ErrorAction SilentlyContinue
            if ($response.StatusCode -eq 200) {
                Write-Host "$Name is ready!" -ForegroundColor Green
                $ready = $true
                break
            }
        } catch {
            Start-Sleep -Seconds 1
            $attempt++
        }
    }
    
    if (-not $ready) {
        Write-Host "$Name failed to start" -ForegroundColor Red
    }
    
    Pop-Location
    return $ready
}

# Start Layer 1: MCP Server
Start-Service -Name "MCP-Server" -Directory "." -Port 8080 -Command "go run cmd/server/main.go"

# Start Layer 2: AI Skin Orchestrator
Start-Service -Name "AI-Skin-Orchestrator" -Directory "ai-skin-orchestrator" -Port 8081 -Command "go run cmd/server/main.go"

# Start Layer 3: Banking Agent
$env:AGENT_TYPE = "BANKING"
$env:SERVER_PORT = "8001"
$env:AGENT_ENDPOINT = "http://localhost:8001"
Start-Service -Name "Banking-Agent" -Directory "agent-mesh" -Port 8001 -Command "go run cmd/server/main.go"

# Start Layer 3: Fraud Agent
$env:AGENT_TYPE = "FRAUD"
$env:SERVER_PORT = "8002"
$env:AGENT_ENDPOINT = "http://localhost:8002"
Start-Service -Name "Fraud-Agent" -Directory "agent-mesh" -Port 8002 -Command "go run cmd/server/main.go"

# Start Layer 5: Banking Integrations
Start-Service -Name "Banking-Integrations" -Directory "banking-integrations" -Port 7000 -Command "go run cmd/server/main.go"

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "Services started! Testing..." -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host ""

# Wait a bit for all services to fully start
Start-Sleep -Seconds 5

# Run tests
Write-Host "Running integration tests..." -ForegroundColor Cyan
Write-Host ""

$API_KEY = "test-api-key"

# Test Layer 1
Write-Host "Testing Layer 1: MCP Server" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
    Write-Host "✓ MCP Server Health: OK" -ForegroundColor Green
} catch {
    Write-Host "✗ MCP Server Health: FAILED" -ForegroundColor Red
}

# Test Layer 2
Write-Host "Testing Layer 2: AI Skin Orchestrator" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8081/health" -Method GET
    Write-Host "✓ Orchestrator Health: OK" -ForegroundColor Green
} catch {
    Write-Host "✗ Orchestrator Health: FAILED" -ForegroundColor Red
}

# Test Layer 3
Write-Host "Testing Layer 3: Agent Mesh" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8001/health" -Method GET
    Write-Host "✓ Banking Agent Health: OK" -ForegroundColor Green
} catch {
    Write-Host "✗ Banking Agent Health: FAILED" -ForegroundColor Red
}

# Test Layer 5
Write-Host "Testing Layer 5: Banking Integrations" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:7000/health" -Method GET
    Write-Host "✓ Banking Integrations Health: OK" -ForegroundColor Green
} catch {
    Write-Host "✗ Banking Integrations Health: FAILED" -ForegroundColor Red
}

# Test End-to-End
Write-Host ""
Write-Host "Testing End-to-End Flow" -ForegroundColor Cyan
Write-Host "Submitting task to MCP Server..." -ForegroundColor Yellow

$taskBody = @{
    user_id = "U10001"
    channel = "MB"
    intent = "TRANSFER_NEFT"
    data = @{
        amount = 50000
        to_account = "XXXX4321"
    }
} | ConvertTo-Json

try {
    $headers = @{
        "Content-Type" = "application/json"
        "X-API-Key" = $API_KEY
    }
    
    $taskResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/submit-task" -Method POST -Body $taskBody -Headers $headers
    $taskId = $taskResponse.task_id
    
    Write-Host "✓ Task submitted: $taskId" -ForegroundColor Green
    
    Start-Sleep -Seconds 3
    
    $resultResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/get-result/$taskId" -Method GET -Headers @{"X-API-Key" = $API_KEY}
    Write-Host "✓ Task completed with status: $($resultResponse.status)" -ForegroundColor Green
    Write-Host "  Risk Score: $($resultResponse.risk_score)" -ForegroundColor Cyan
    Write-Host "  Explanation: $($resultResponse.explanation)" -ForegroundColor Cyan
    
} catch {
    Write-Host "✗ End-to-End test failed: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "Testing Complete!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Services running on:" -ForegroundColor Cyan
Write-Host "  - MCP Server: http://localhost:8080"
Write-Host "  - AI Skin Orchestrator: http://localhost:8081"
Write-Host "  - Banking Agent: http://localhost:8001"
Write-Host "  - Fraud Agent: http://localhost:8002"
Write-Host "  - Banking Integrations: http://localhost:7000"
Write-Host ""
Write-Host "Logs are in the 'logs/' directory" -ForegroundColor Yellow
Write-Host "Press Ctrl+C to stop all services" -ForegroundColor Yellow

