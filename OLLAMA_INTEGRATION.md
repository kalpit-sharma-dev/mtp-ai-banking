# Ollama/Llama 3 Integration Guide

## Overview

The AI Banking Platform now supports **Ollama** with **Llama 3** as an alternative to ChatGPT. This allows you to run the LLM locally in Docker, providing better privacy and cost control.

## Features

✅ **Ollama Integration**: Full support for Ollama API  
✅ **Session Management**: Conversation history is stored per session  
✅ **Typing Effect**: ChatGPT-like character-by-character typing animation in UI  
✅ **Fallback Support**: Can still use OpenAI if Ollama is unavailable  

## Setup

### 1. Install and Run Ollama

#### Option A: Docker (Recommended)

```bash
# Pull Ollama Docker image
docker pull ollama/ollama

# Run Ollama container
docker run -d -p 11434:11434 --name ollama ollama/ollama

# Pull Llama 3 model
docker exec -it ollama ollama pull llama3
```

#### Option B: Native Installation

```bash
# Install Ollama (macOS/Linux)
curl -fsSL https://ollama.com/install.sh | sh

# Pull Llama 3 model
ollama pull llama3
```

### 2. Verify Ollama is Running

```bash
# Check if Ollama is accessible
curl http://localhost:11434/api/tags
```

### 3. Configure Environment Variables

Update `.env` file in `ai-skin-orchestrator/`:

```env
# LLM Configuration
LLM_PROVIDER=ollama
LLM_MODEL=llama3
LLM_BASE_URL=http://localhost:11434
LLM_ENABLED=true
LLM_TEMPERATURE=0.7
LLM_MAX_TOKENS=1000
```

Or use OpenAI:

```env
LLM_PROVIDER=openai
LLM_MODEL=gpt-3.5-turbo
LLM_API_KEY=your-openai-api-key
LLM_ENABLED=true
```

## Architecture

### Components

1. **OllamaService** (`ai-skin-orchestrator/internal/service/ollama_service.go`)
   - Handles Ollama API communication
   - Supports streaming responses
   - Builds prompts with conversation history

2. **SessionService** (`ai-skin-orchestrator/internal/service/session_service.go`)
   - Manages conversation sessions
   - Stores message history (last 20 messages)
   - 24-hour TTL for sessions

3. **LLMService** (`ai-skin-orchestrator/internal/service/llm_service.go`)
   - Unified interface for both Ollama and OpenAI
   - Automatically routes to correct provider
   - Supports conversation history

4. **UI Typing Effect** (`web-ui/src/pages/AIAssistant.jsx`)
   - Character-by-character typing animation
   - Blinking cursor during typing
   - Smooth scrolling to latest message

## Session Management

### How It Works

1. **First Request**: Creates a new session automatically
2. **Subsequent Requests**: Uses existing session ID from response
3. **Conversation History**: Last 20 messages stored per session
4. **Session Expiry**: 24 hours of inactivity

### Session Flow

```
User Request
    ↓
Orchestrator Controller
    ↓
SessionService.GetOrCreateSession()
    ↓
Add User Message to Session
    ↓
Process Request (with history)
    ↓
Add Bot Response to Session
    ↓
Return Response (with session_id)
```

## Typing Effect

### Implementation

The UI implements a character-by-character typing effect:

- **Speed**: 20ms per character
- **Cursor**: Animated blinking cursor during typing
- **Completion**: Cursor disappears when typing completes

### Code Example

```javascript
const typeMessage = (messageObj, fullText, messageIndex) => {
  let currentIndex = 0
  const typingSpeed = 20 // milliseconds per character

  const typeChar = () => {
    if (currentIndex < fullText.length) {
      const newContent = fullText.substring(0, currentIndex + 1)
      // Update message content
      currentIndex++
      setTimeout(typeChar, typingSpeed)
    }
  }
  typeChar()
}
```

## Testing

### Test Ollama Connection

```bash
curl http://localhost:11434/api/generate -d '{
  "model": "llama3",
  "prompt": "Hello, how are you?",
  "stream": false
}'
```

### Test with Banking Request

```bash
curl -X POST http://localhost:8081/api/v1/process \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "user_id": "U10001",
    "channel": "MB",
    "input": "Check my balance",
    "input_type": "natural_language"
  }'
```

## Troubleshooting

### Ollama Not Responding

1. **Check if Ollama is running**:
   ```bash
   docker ps | grep ollama
   # or
   ps aux | grep ollama
   ```

2. **Check Ollama logs**:
   ```bash
   docker logs ollama
   ```

3. **Verify model is downloaded**:
   ```bash
   docker exec -it ollama ollama list
   ```

### Model Not Found

If you get "model not found" error:

```bash
# Pull the model
docker exec -it ollama ollama pull llama3

# Or for native installation
ollama pull llama3
```

### Session Issues

- Sessions are stored in-memory (will be lost on restart)
- For production, consider using Redis for session persistence
- Session TTL is 24 hours by default

## Performance Considerations

### Ollama vs OpenAI

- **Ollama**: Slower but free, runs locally
- **OpenAI**: Faster but costs money, requires internet

### Optimization Tips

1. **Use smaller models** for faster responses (e.g., `llama3:8b` instead of `llama3:70b`)
2. **Adjust temperature** for faster/slower responses
3. **Limit conversation history** (currently 20 messages max)

## Future Enhancements

- [ ] Redis-based session persistence
- [ ] Streaming responses in real-time (SSE/WebSocket)
- [ ] Multiple model support (switch between models)
- [ ] Model fine-tuning for banking domain
- [ ] Response caching for common queries

## Configuration Reference

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_PROVIDER` | `ollama` | LLM provider: `ollama` or `openai` |
| `LLM_MODEL` | `llama3` | Model name |
| `LLM_BASE_URL` | `http://localhost:11434` | Ollama server URL |
| `LLM_API_KEY` | - | OpenAI API key (if using OpenAI) |
| `LLM_ENABLED` | `true` | Enable/disable LLM |
| `LLM_TEMPERATURE` | `0.7` | Response creativity (0.0-1.0) |
| `LLM_MAX_TOKENS` | `1000` | Maximum response length |

## Support

For issues or questions:
1. Check Ollama logs: `docker logs ollama`
2. Check orchestrator logs for LLM errors
3. Verify model is downloaded and accessible
4. Test Ollama API directly with curl

