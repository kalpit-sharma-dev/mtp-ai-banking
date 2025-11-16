package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aibanking/ai-skin-orchestrator/internal/config"
	"github.com/rs/zerolog/log"
)

// OllamaService handles interactions with Ollama/Llama models
type OllamaService struct {
	baseURL        string
	embeddingURL   string
	httpClient     *http.Client
	model          string
	embeddingModel string // Model for embeddings (can be different from chat model)
	enabled        bool
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// OllamaResponse represents a streaming response from Ollama
type OllamaResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	DoneReason         string `json:"done_reason,omitempty"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

// StreamingSession represents a streaming response session
type StreamingSession struct {
	ID        string
	Content   string
	Done      bool
	CreatedAt time.Time
	mu        interface{} // For future thread-safety if needed
}

// NewOllamaService creates a new Ollama service
func NewOllamaService(cfg *config.LLMConfig) *OllamaService {
	if !cfg.Enabled || cfg.BaseURL == "" {
		log.Info().Msg("Ollama service disabled or base URL not provided")
		return &OllamaService{
			enabled: false,
		}
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	embeddingBaseURL := baseURL
	
	// Set up generate endpoint
	if !strings.HasSuffix(baseURL, "/api/generate") {
		baseURL = baseURL + "/api/generate"
	}
	
	// Set up embeddings endpoint
	if !strings.HasSuffix(embeddingBaseURL, "/api/embed") {
		embeddingBaseURL = strings.TrimSuffix(embeddingBaseURL, "/api/generate")
		embeddingBaseURL = embeddingBaseURL + "/api/embed"
	}
	
	// Default embedding model (can use nomic-embed-text or mxbai-embed-large)
	embeddingModel := "nomic-embed-text" // Good default for embeddings
	
	return &OllamaService{
		baseURL:        baseURL,
		embeddingURL:   embeddingBaseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Increase timeout for longer responses
		},
		model:          cfg.Model,
		embeddingModel: embeddingModel,
		enabled:        true,
	}
}

// CallLLM calls Ollama with a prompt and returns the complete response
func (os *OllamaService) CallLLM(ctx context.Context, prompt string) (string, error) {
	if !os.enabled {
		return "", fmt.Errorf("Ollama service is disabled")
	}

	session := &StreamingSession{
		ID:        fmt.Sprintf("session_%d", time.Now().UnixNano()),
		Content:   "",
		Done:      false,
		CreatedAt: time.Now(),
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// Start streaming
	os.QueryStreaming(ctx, prompt, session)

	// Wait for completion or timeout
	for !session.Done {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return session.Content, fmt.Errorf("request timed out after 2 minutes")
			}
			return session.Content, ctx.Err()
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	return session.Content, nil
}

// QueryStreaming streams responses from Ollama
func (os *OllamaService) QueryStreaming(ctx context.Context, prompt string, session *StreamingSession) {
	log.Info().Int("prompt_length", len(prompt)).Msg("Starting Ollama streaming request")

	// Prepare the request with banking-specific options
	requestBody := OllamaRequest{
		Model:  os.model,
		Prompt: prompt,
		Stream: true,
		Options: map[string]interface{}{
			"temperature":    0.7,
			"top_p":         0.9,
			"top_k":         40,
			"num_predict":   1000,
			"stop":          []string{"Human:", "User:"},
			"repeat_penalty": 1.1,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling request body")
		session.Content = fmt.Sprintf("Error: Failed to prepare request - %v", err)
		session.Done = true
		return
	}

	// Create the request with context
	req, err := http.NewRequestWithContext(ctx, "POST", os.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Error().Err(err).Msg("Error creating request")
		session.Content = fmt.Sprintf("Error: Failed to create request - %v", err)
		session.Done = true
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	log.Info().Str("url", os.baseURL).Msg("Sending request to Ollama")

	// Send the request
	resp, err := os.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Error sending request")
		if strings.Contains(err.Error(), "connection refused") {
			session.Content = "Error: Cannot connect to Ollama service. Please ensure Ollama is running."
		} else {
			session.Content = fmt.Sprintf("Error: Network error - %v", err)
		}
		session.Done = true
		return
	}
	defer resp.Body.Close()

	log.Info().Str("status", resp.Status).Msg("Received response from Ollama")

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Error().Str("status", resp.Status).Str("body", string(body)).Msg("Error from Ollama API")
		if resp.StatusCode == 404 {
			session.Content = "Error: Ollama model not found. Please ensure the model is downloaded."
		} else if resp.StatusCode == 500 {
			session.Content = "Error: Ollama server error. Please check the server logs."
		} else {
			session.Content = fmt.Sprintf("Error: API returned status %s", resp.Status)
		}
		session.Done = true
		return
	}

	// Process the streaming response
	os.processStreamingResponse(ctx, session, resp)
}

// processStreamingResponse processes the streaming response from Ollama
func (os *OllamaService) processStreamingResponse(ctx context.Context, session *StreamingSession, resp *http.Response) {
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			session.Done = true
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					session.Done = true
					return
				}
				log.Error().Err(err).Msg("Error reading response")
				session.Done = true
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Parse the JSON response
			var ollamaResp OllamaResponse
			if err := json.Unmarshal([]byte(line), &ollamaResp); err != nil {
				log.Warn().Err(err).Str("line", line).Msg("Error parsing Ollama response")
				continue
			}

			// Append response chunk
			if ollamaResp.Response != "" {
				session.Content += ollamaResp.Response
			}

			// If the response is done, mark the session as complete
			if ollamaResp.Done {
				session.Done = true
				session.Content = strings.TrimSpace(session.Content)
				return
			}
		}
	}
}

// BuildPromptWithContext builds a prompt with conversation history
func (os *OllamaService) BuildPromptWithContext(message string, conversationHistory []map[string]string) string {
	var promptBuilder strings.Builder

	const bankingSystemPrompt = `You are a secure and intelligent AI banking assistant integrated into a digital banking system.

Your role is to help users perform a wide range of banking tasks safely, efficiently, and clearly. Always ensure user intent is well-understood, confirm sensitive operations, and provide helpful, accurate guidance at every step.

**IMPORTANT: Always identify yourself as a banking assistant in your responses, especially when responding to greetings or questions about your capabilities.**

You have access to the following banking functions:
- fund_transfer: Transfer funds to a saved payee. Confirm the recipient name and amount before initiating the transaction.
- add_payee: Add a new payee with details like name, account number, and IFSC. Ensure confirmation before saving.
- view_balance: Provide the current account balance on request.
- get_statement: Retrieve account statements and transaction history.
- create_fd: Create a fixed deposit by specifying amount and duration.
- apply_loan: Help users apply for loans and check eligibility.

**Response Guidelines:**
- **For greetings** (hello, hi, how are you): Greet the user warmly and clearly identify yourself as their AI banking assistant. Ask how you can help them with their banking needs.
- **For capability questions** (what can you do, what operations do you support): Clearly list all the banking operations you can help with, including balance checks, fund transfers, statements, beneficiary management, fixed deposits, loans, and credit scores.
- **For banking operations**: Be professional, concise, and user-friendly. Always maintain security and confidentiality.
- **For confirmations**: Confirm high-risk operations like fund_transfer or add_payee before execution.
- **For guidance**: If a user seems unsure, guide them step-by-step.

Always remember: You are a BANKING ASSISTANT. Make this clear in your responses, especially when users greet you or ask about your capabilities.

You are here to make banking simpler, safer, and smarter for the user.`

	promptBuilder.WriteString(bankingSystemPrompt)

	// Add conversation history if available
	if len(conversationHistory) > 0 {
		promptBuilder.WriteString("\n\nConversation history:\n")
		// Include last 6 messages (3 exchanges) for context
		start := 0
		if len(conversationHistory) > 6 {
			start = len(conversationHistory) - 6
		}
		for i := start; i < len(conversationHistory); i++ {
			msg := conversationHistory[i]
			role := "Human"
			if msg["role"] == "assistant" || msg["role"] == "bot" {
				role = "Assistant"
			}
			promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", role, msg["content"]))
		}
		promptBuilder.WriteString("\n")
	}

	// Add current message
	promptBuilder.WriteString(fmt.Sprintf("Human: %s\n", message))
	promptBuilder.WriteString("Assistant:")

	return promptBuilder.String()
}

// OllamaEmbeddingRequest represents a request to Ollama embeddings API
type OllamaEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// OllamaEmbeddingResponse represents a response from Ollama embeddings API
type OllamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// GenerateEmbedding generates embeddings using Ollama's embedding API
func (os *OllamaService) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if !os.enabled {
		return nil, fmt.Errorf("Ollama service is disabled")
	}

	// Prepare embedding request
	requestBody := OllamaEmbeddingRequest{
		Model: os.embeddingModel,
		Input: text,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	// Create request with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", os.embeddingURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	log.Debug().
		Str("url", os.embeddingURL).
		Str("model", os.embeddingModel).
		Int("text_length", len(text)).
		Msg("Requesting embeddings from Ollama")

	// Send request
	resp, err := os.httpClient.Do(req)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get embeddings from Ollama, will use TF-IDF fallback")
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Warn().
			Int("status", resp.StatusCode).
			Str("response", string(bodyBytes)).
			Msg("Ollama embeddings API returned error")
		return nil, fmt.Errorf("embedding API returned status %d", resp.StatusCode)
	}

	// Parse response
	var embeddingResp OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	if len(embeddingResp.Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	log.Debug().
		Int("embedding_dim", len(embeddingResp.Embedding)).
		Msg("Successfully generated embeddings from Ollama")

	return embeddingResp.Embedding, nil
}

