# Quick Start Guide - MCP Server

## Prerequisites

- Go 1.21 or higher installed
- Redis server (or Docker for Redis)
- Basic understanding of REST APIs

## Step 1: Install Dependencies

```bash
go mod download
```

## Step 2: Setup Environment

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env if needed (defaults should work for local development)
```

## Step 3: Start Redis

**Option A: Using Docker (Recommended)**
```bash
docker run -d -p 6379:6379 --name redis redis:latest
```

**Option B: Using Local Redis**
```bash
redis-server
```

## Step 4: Run the Server

```bash
# Using Make
make run

# OR directly with Go
go run cmd/server/main.go
```

You should see:
```
{"level":"info","message":"Starting MCP Server for AI Banking Platform"}
{"level":"info","message":"Connected to Redis"}
{"level":"info","address":"0.0.0.0:8080","message":"MCP Server started"}
```

## Step 5: Test the API

### Using the Test Script

```bash
# Make sure jq is installed for JSON formatting
# On Windows: Install via chocolatey or use WSL
# On Linux/Mac: sudo apt-get install jq / brew install jq

./examples/test_api.sh
```

### Manual Testing with cURL

**1. Health Check**
```bash
curl http://localhost:8080/health
```

**2. Submit a Task**
```bash
curl -X POST http://localhost:8080/api/v1/submit-task \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "user_id": "U10001",
    "channel": "MB",
    "intent": "TRANSFER_NEFT",
    "data": {
      "amount": 50000,
      "to_account": "XXXX4321"
    }
  }'
```

**3. Get Task Result** (use task_id from previous response)
```bash
curl -X GET http://localhost:8080/api/v1/get-result/task_<your-task-id> \
  -H "X-API-Key: test-api-key"
```

## Common Issues

### Redis Connection Error
```
Failed to connect to Redis, continuing without persistence
```
**Solution**: Make sure Redis is running on `localhost:6379` or update `.env` with correct Redis host/port.

### Port Already in Use
```
bind: address already in use
```
**Solution**: Change `SERVER_PORT` in `.env` or stop the process using port 8080.

### Module Not Found
```
cannot find module
```
**Solution**: Run `go mod download` and `go mod tidy`.

## Next Steps

1. Review the API documentation in `README.md`
2. Check `LAYER_1_SUMMARY.md` for implementation details
3. Explore the code structure in `internal/` directory
4. Customize routing rules via `/api/v1/rules/upload`
5. Register your own agents via `/api/v1/register-agent`

## Using Docker Compose

For a complete setup with Redis:

```bash
docker-compose up
```

This will start both Redis and the MCP Server together.

## Development

### Build Binary
```bash
make build
# Binary will be in bin/mcp-server
```

### Run Tests
```bash
make test
```

### Format Code
```bash
make fmt
```

## Support

For issues or questions, refer to:
- `README.md` - Full documentation
- `LAYER_1_SUMMARY.md` - Implementation details
- `project document.txt` - Overall project architecture

