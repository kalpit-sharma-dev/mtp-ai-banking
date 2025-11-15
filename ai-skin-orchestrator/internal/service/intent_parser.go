package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/aibanking/ai-skin-orchestrator/internal/model"
	"github.com/rs/zerolog/log"
)

// IntentParser parses user input to extract intent and entities
type IntentParser struct {
	llmService *LLMService
	useLLM    bool
}

// NewIntentParser creates a new intent parser
func NewIntentParser(llmService *LLMService, useLLM bool) *IntentParser {
	return &IntentParser{
		llmService: llmService,
		useLLM:    useLLM,
	}
}

// ParseIntent parses user input to extract intent and entities
func (ip *IntentParser) ParseIntent(ctx context.Context, userInput string, inputType string) (*model.Intent, error) {
	if inputType == "structured" {
		// If structured, extract directly
		return ip.parseStructuredInput(userInput)
	}

	// For natural language, use LLM if available, otherwise use rule-based
	if ip.useLLM && ip.llmService != nil {
		return ip.parseWithLLM(ctx, userInput)
	}

	return ip.parseWithRules(userInput)
}

// parseStructuredInput parses structured JSON input
func (ip *IntentParser) parseStructuredInput(input string) (*model.Intent, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}

	intentType, ok := data["intent"].(string)
	if !ok {
		return nil, fmt.Errorf("missing intent in structured input")
	}

	entities := make(map[string]interface{})
	if entitiesData, ok := data["entities"].(map[string]interface{}); ok {
		entities = entitiesData
	}

	return &model.Intent{
		Type:        model.IntentType(intentType),
		Confidence:  1.0,
		Entities:    entities,
		OriginalText: input,
	}, nil
}

// parseWithLLM uses LLM to parse natural language intent
func (ip *IntentParser) parseWithLLM(ctx context.Context, userInput string) (*model.Intent, error) {
	prompt := fmt.Sprintf(`Analyze the following banking request and extract:
1. Intent type (one of: TRANSFER_NEFT, TRANSFER_RTGS, TRANSFER_IMPS, TRANSFER_UPI, CHECK_BALANCE, GET_STATEMENT, ADD_BENEFICIARY, APPLY_LOAN, CREDIT_SCORE)
2. Entities (amount, account number, beneficiary, etc.)
3. Confidence score (0.0 to 1.0)

User request: "%s"

Respond in JSON format:
{
  "intent": "INTENT_TYPE",
  "confidence": 0.95,
  "entities": {
    "amount": 50000,
    "to_account": "XXXX4321"
  }
}`, userInput)

	response, err := ip.llmService.CallLLM(ctx, prompt)
	if err != nil {
		log.Warn().Err(err).Msg("LLM parsing failed, falling back to rules")
		return ip.parseWithRules(userInput)
	}

	var result struct {
		Intent     string                 `json:"intent"`
		Confidence float64                `json:"confidence"`
		Entities   map[string]interface{} `json:"entities"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return ip.parseWithRules(userInput)
	}

	return &model.Intent{
		Type:        model.IntentType(result.Intent),
		Confidence:  result.Confidence,
		Entities:   result.Entities,
		OriginalText: userInput,
	}, nil
}

// parseWithRules uses rule-based parsing for natural language
func (ip *IntentParser) parseWithRules(userInput string) (*model.Intent, error) {
	input := strings.ToLower(userInput)
	entities := make(map[string]interface{})

	// Extract amount
	amountRegex := regexp.MustCompile(`(?i)(?:rs\.?|₹|rupees?)?\s*(\d+(?:,\d{3})*(?:\.\d{2})?)`)
	if matches := amountRegex.FindStringSubmatch(input); len(matches) > 1 {
		amountStr := strings.ReplaceAll(matches[1], ",", "")
		entities["amount"] = amountStr
	}

	// Extract account number
	accountRegex := regexp.MustCompile(`(?i)(?:account|acc|ac)\s*(?:no|number|#)?\s*:?\s*([\dX]{4,})`)
	if matches := accountRegex.FindStringSubmatch(input); len(matches) > 1 {
		entities["to_account"] = matches[1]
	}

	// Extract IFSC
	ifscRegex := regexp.MustCompile(`(?i)ifsc\s*:?\s*([A-Z]{4}0[A-Z0-9]{6})`)
	if matches := ifscRegex.FindStringSubmatch(input); len(matches) > 1 {
		entities["ifsc"] = matches[1]
	}

	// Determine intent based on keywords
	var intentType model.IntentType
	var confidence float64 = 0.7

	switch {
	case containsAny(input, []string{"neft", "transfer neft", "send via neft", "transfer", "send money", "pay"}):
		intentType = model.IntentTransferNEFT
		confidence = 0.9
	case containsAny(input, []string{"rtgs", "transfer rtgs"}):
		intentType = model.IntentTransferRTGS
		confidence = 0.9
	case containsAny(input, []string{"imps", "transfer imps"}):
		intentType = model.IntentTransferIMPS
		confidence = 0.9
	case containsAny(input, []string{"upi", "pay via upi", "scan qr"}):
		intentType = model.IntentTransferUPI
		confidence = 0.9
	case containsAny(input, []string{"balance", "check balance", "account balance", "how much", "what is my balance"}):
		intentType = model.IntentCheckBalance
		confidence = 0.95
	case containsAny(input, []string{"statement", "mini statement", "transaction history", "transactions", "history"}):
		intentType = model.IntentGetStatement
		confidence = 0.9
	case containsAny(input, []string{"add beneficiary", "add payee", "save beneficiary", "beneficiary"}):
		intentType = model.IntentAddBeneficiary
		confidence = 0.9
	case containsAny(input, []string{"loan", "apply loan", "personal loan"}):
		intentType = model.IntentApplyLoan
		confidence = 0.85
	case containsAny(input, []string{"credit score", "cibil score", "credit rating"}):
		intentType = model.IntentCreditScore
		confidence = 0.85
	default:
		intentType = model.IntentUnknown
		confidence = 0.3
	}

	return &model.Intent{
		Type:        intentType,
		Confidence:  confidence,
		Entities:    entities,
		OriginalText: userInput,
	}, nil
}

func containsAny(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(s, keyword) {
			return true
		}
	}
	return false
}

