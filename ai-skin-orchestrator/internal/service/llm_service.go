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
	client    *openai.Client
	enabled   bool
	model     string
	temperature float64
	maxTokens int
}

// NewLLMService creates a new LLM service
func NewLLMService(cfg *config.LLMConfig) *LLMService {
	if !cfg.Enabled || cfg.APIKey == "" {
		log.Info().Msg("LLM service disabled or API key not provided")
		return &LLMService{
			enabled: false,
		}
	}

	var client *openai.Client
	if cfg.BaseURL != "" {
		// Custom base URL for self-hosted models
		config := openai.DefaultConfig(cfg.APIKey)
		config.BaseURL = cfg.BaseURL
		client = openai.NewClientWithConfig(config)
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
	}
}

// CallLLM calls the LLM with a prompt and returns the response
func (ls *LLMService) CallLLM(ctx context.Context, prompt string) (string, error) {
	if !ls.enabled {
		return "", fmt.Errorf("LLM service is disabled")
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

