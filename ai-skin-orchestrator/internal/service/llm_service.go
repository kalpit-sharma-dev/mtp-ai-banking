package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aibanking/ai-skin-orchestrator/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

// LLMService handles LLM interactions for intent parsing and natural language understanding
type LLMService struct {
	client      *openai.Client
	ollama      *OllamaService
	enabled     bool
	model       string
	temperature float64
	maxTokens   int
	provider    string // "openai" or "ollama"
}

// NewLLMService creates a new LLM service
func NewLLMService(cfg *config.LLMConfig) *LLMService {
	if !cfg.Enabled {
		log.Info().Msg("LLM service disabled")
		return &LLMService{
			enabled: false,
		}
	}

	provider := strings.ToLower(cfg.Provider)
	if provider == "ollama" {
		// Use Ollama service
		ollama := NewOllamaService(cfg)
		return &LLMService{
			ollama:      ollama,
			enabled:     true,
			model:       cfg.Model,
			temperature: cfg.Temperature,
			maxTokens:   cfg.MaxTokens,
			provider:    "ollama",
		}
	}

	// Default to OpenAI
	if cfg.APIKey == "" {
		log.Info().Msg("LLM service disabled - API key not provided")
		return &LLMService{
			enabled: false,
		}
	}

	var client *openai.Client
	if cfg.BaseURL != "" {
		// Custom base URL for self-hosted models
		openaiConfig := openai.DefaultConfig(cfg.APIKey)
		openaiConfig.BaseURL = cfg.BaseURL
		client = openai.NewClientWithConfig(openaiConfig)
	} else {
		// Standard OpenAI
		client = openai.NewClient(cfg.APIKey)
	}

	return &LLMService{
		client:      client,
		enabled:     true,
		model:       cfg.Model,
		temperature: cfg.Temperature,
		maxTokens:   cfg.MaxTokens,
		provider:    "openai",
	}
}

// CallLLM calls the LLM with a prompt and returns the response
func (ls *LLMService) CallLLM(ctx context.Context, prompt string) (string, error) {
	if !ls.enabled {
		return "", fmt.Errorf("LLM service is disabled")
	}

	// Use Ollama if configured
	if ls.provider == "ollama" && ls.ollama != nil {
		return ls.ollama.CallLLM(ctx, prompt)
	}

	// Default to OpenAI
	if ls.client == nil {
		return "", fmt.Errorf("LLM client not initialized")
	}

	resp, err := ls.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: ls.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: float32(ls.temperature),
			MaxTokens:   ls.maxTokens,
		},
	)

	if err != nil {
		return "", fmt.Errorf("LLM API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	
	// Try to extract JSON if response is wrapped
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	return content, nil
}

// CallLLMWithHistory calls the LLM with conversation history (for Ollama)
func (ls *LLMService) CallLLMWithHistory(ctx context.Context, message string, conversationHistory []map[string]string) (string, error) {
	if !ls.enabled {
		return "", fmt.Errorf("LLM service is disabled")
	}

	// Use Ollama if configured
	if ls.provider == "ollama" && ls.ollama != nil {
		prompt := ls.ollama.BuildPromptWithContext(message, conversationHistory)
		return ls.ollama.CallLLM(ctx, prompt)
	}

	// For OpenAI, build messages from history
	if ls.client == nil {
		return "", fmt.Errorf("LLM client not initialized")
	}

	messages := []openai.ChatCompletionMessage{}
	
	// Add conversation history
	for _, msg := range conversationHistory {
		role := openai.ChatMessageRoleUser
		if msg["role"] == "assistant" || msg["role"] == "bot" {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg["content"],
		})
	}
	
	// Add current message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	resp, err := ls.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       ls.model,
			Messages:    messages,
			Temperature: float32(ls.temperature),
			MaxTokens:   ls.maxTokens,
		},
	)

	if err != nil {
		return "", fmt.Errorf("LLM API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// ParseIntentWithLLM uses LLM to parse natural language intent
func (ls *LLMService) ParseIntentWithLLM(ctx context.Context, userInput string) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`You are a banking assistant. Analyze the following user request and extract:
1. Intent type (one of: TRANSFER_NEFT, TRANSFER_RTGS, TRANSFER_IMPS, TRANSFER_UPI, CHECK_BALANCE, GET_STATEMENT, ADD_BENEFICIARY, APPLY_LOAN, CREDIT_SCORE)
2. Entities (amount, account number, beneficiary name, IFSC code, etc.)
3. Confidence score (0.0 to 1.0)

User request: "%s"

Respond ONLY with valid JSON in this format:
{
  "intent": "INTENT_TYPE",
  "confidence": 0.95,
  "entities": {
    "amount": 50000,
    "to_account": "XXXX4321",
    "ifsc": "BANK0001234"
  }
}`, userInput)

	response, err := ls.CallLLM(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return result, nil
}

