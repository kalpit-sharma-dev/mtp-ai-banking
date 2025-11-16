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

// IntentPattern represents a pattern for intent recognition
type IntentPattern struct {
	Name       string
	Patterns   []*regexp.Regexp
	Keywords   map[string]float64
	IntentType model.IntentType
}

// IntentParser parses user input to extract intent and entities
type IntentParser struct {
	llmService *LLMService
	useLLM     bool
	patterns   []IntentPattern
}

// NewIntentParser creates a new intent parser
func NewIntentParser(llmService *LLMService, useLLM bool) *IntentParser {
	parser := &IntentParser{
		llmService: llmService,
		useLLM:     useLLM,
	}
	parser.initializeIntentPatterns()
	return parser
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
		Type:         model.IntentType(intentType),
		Confidence:   1.0,
		Entities:     entities,
		OriginalText: input,
	}, nil
}

// parseWithLLM uses LLM to parse natural language intent
func (ip *IntentParser) parseWithLLM(ctx context.Context, userInput string) (*model.Intent, error) {
	prompt := fmt.Sprintf(`Analyze the following banking request and extract:
1. Intent type (one of: TRANSFER_NEFT, TRANSFER_RTGS, TRANSFER_IMPS, TRANSFER_UPI, CHECK_BALANCE, GET_STATEMENT, ADD_BENEFICIARY, APPLY_LOAN, CREDIT_SCORE, CONVERSATIONAL)
2. Entities (amount, account number, beneficiary, etc.)
3. Confidence score (0.0 to 1.0)

Note: If the request is a greeting (hello, hi, how are you), question about capabilities (what can you do, what operations do you support), or a conversational query, use intent: "CONVERSATIONAL".

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
		Type:         model.IntentType(result.Intent),
		Confidence:   result.Confidence,
		Entities:     result.Entities,
		OriginalText: userInput,
	}, nil
}

// initializeIntentPatterns initializes intent patterns with regex and weighted keywords
func (ip *IntentParser) initializeIntentPatterns() {
	ip.patterns = []IntentPattern{
		{
			Name: "fund_transfer_neft",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)transfer\s+(\d+(?:,\d{3})*(?:\.\d{2})?)\s+(?:rs\.?|₹|rupees?)?\s*(?:via\s+)?neft`),
				regexp.MustCompile(`(?i)send\s+(\d+(?:,\d{3})*(?:\.\d{2})?)\s+(?:rs\.?|₹|rupees?)?\s*(?:via\s+)?neft`),
				regexp.MustCompile(`(?i)neft\s+transfer\s+of\s+(\d+(?:,\d{3})*(?:\.\d{2})?)`),
			},
			Keywords: map[string]float64{
				"neft":     1.0,
				"transfer": 0.9,
				"send":     0.8,
				"money":    0.7,
				"amount":   0.6,
			},
			IntentType: model.IntentTransferNEFT,
		},
		{
			Name: "fund_transfer_rtgs",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)transfer\s+(\d+(?:,\d{3})*(?:\.\d{2})?)\s+(?:rs\.?|₹|rupees?)?\s*(?:via\s+)?rtgs`),
				regexp.MustCompile(`(?i)send\s+(\d+(?:,\d{3})*(?:\.\d{2})?)\s+(?:rs\.?|₹|rupees?)?\s*(?:via\s+)?rtgs`),
				regexp.MustCompile(`(?i)rtgs\s+transfer\s+of\s+(\d+(?:,\d{3})*(?:\.\d{2})?)`),
			},
			Keywords: map[string]float64{
				"rtgs":     1.0,
				"transfer": 0.9,
				"send":     0.8,
				"money":    0.7,
			},
			IntentType: model.IntentTransferRTGS,
		},
		{
			Name: "fund_transfer_imps",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)transfer\s+(\d+(?:,\d{3})*(?:\.\d{2})?)\s+(?:rs\.?|₹|rupees?)?\s*(?:via\s+)?imps`),
				regexp.MustCompile(`(?i)send\s+(\d+(?:,\d{3})*(?:\.\d{2})?)\s+(?:rs\.?|₹|rupees?)?\s*(?:via\s+)?imps`),
				regexp.MustCompile(`(?i)imps\s+transfer\s+of\s+(\d+(?:,\d{3})*(?:\.\d{2})?)`),
			},
			Keywords: map[string]float64{
				"imps":     1.0,
				"transfer": 0.9,
				"send":     0.8,
				"money":    0.7,
			},
			IntentType: model.IntentTransferIMPS,
		},
		{
			Name: "fund_transfer_upi",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)transfer\s+(\d+(?:,\d{3})*(?:\.\d{2})?)\s+(?:rs\.?|₹|rupees?)?\s*(?:via\s+)?upi`),
				regexp.MustCompile(`(?i)send\s+(\d+(?:,\d{3})*(?:\.\d{2})?)\s+(?:rs\.?|₹|rupees?)?\s*(?:via\s+)?upi`),
				regexp.MustCompile(`(?i)pay\s+(\d+(?:,\d{3})*(?:\.\d{2})?)\s+(?:rs\.?|₹|rupees?)?\s*(?:via\s+)?upi`),
				regexp.MustCompile(`(?i)upi\s+(?:to|for)\s+([a-zA-Z0-9@._-]+)`),
			},
			Keywords: map[string]float64{
				"upi":      1.0,
				"transfer": 0.9,
				"send":     0.8,
				"pay":      0.8,
				"money":    0.7,
			},
			IntentType: model.IntentTransferUPI,
		},
		{
			Name: "check_balance",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)balance\s+(?:of|for|in)\s+(?:my\s+)?account\s*(\d*)`),
				regexp.MustCompile(`(?i)how\s+much\s+(?:do\s+I\s+have|is\s+in\s+my\s+account|money\s+do\s+I\s+have)`),
				regexp.MustCompile(`(?i)what\s+is\s+my\s+(?:account\s+)?balance`),
				regexp.MustCompile(`(?i)show\s+my\s+balance`),
				regexp.MustCompile(`(?i)check\s+(?:my\s+)?(?:account\s+)?balance`),
			},
			Keywords: map[string]float64{
				"balance":   1.0,
				"amount":    0.8,
				"available": 0.7,
				"check":     0.6,
				"show":      0.6,
				"how much":  0.7,
			},
			IntentType: model.IntentCheckBalance,
		},
		{
			Name: "get_statement",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)statement\s+(?:of|for)\s+(?:my\s+)?account`),
				regexp.MustCompile(`(?i)mini\s+statement`),
				regexp.MustCompile(`(?i)transaction\s+history`),
				regexp.MustCompile(`(?i)show\s+my\s+transactions`),
				regexp.MustCompile(`(?i)recent\s+transactions`),
			},
			Keywords: map[string]float64{
				"statement":    1.0,
				"transactions": 0.9,
				"history":      0.8,
				"mini":         0.7,
				"recent":       0.6,
			},
			IntentType: model.IntentGetStatement,
		},
		{
			Name: "query_last_transaction",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)(?:what|which|show)\s+(?:was|is)\s+(?:my\s+)?(?:last|previous|recent)\s+transaction`),
				regexp.MustCompile(`(?i)(?:what|which|show)\s+(?:was|is)\s+(?:the\s+)?(?:last|previous|recent)\s+transaction`),
				regexp.MustCompile(`(?i)(?:last|previous|recent)\s+transaction`),
				regexp.MustCompile(`(?i)what\s+(?:did\s+I\s+)?(?:transfer|send|pay)\s+(?:last|recently|recent)`),
				regexp.MustCompile(`(?i)my\s+(?:last|previous|recent)\s+transaction`),
				regexp.MustCompile(`(?i)what\s+(?:was|is)\s+(?:my\s+)?(?:last|previous)\s+(?:transfer|payment|transaction)`),
			},
			Keywords: map[string]float64{
				"last":        1.0,
				"previous":    0.9,
				"recent":      0.8,
				"transaction": 0.9,
				"transfer":    0.8,
				"what":        0.7,
			},
			IntentType: model.IntentConversational, // Route to conversational for RAG to handle
		},
		{
			Name: "add_beneficiary",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)add\s+(?:new\s+)?(?:payee|beneficiary)\s+(?:named\s+)?([a-zA-Z\s]+)`),
				regexp.MustCompile(`(?i)save\s+(?:new\s+)?(?:payee|beneficiary)\s+(?:named\s+)?([a-zA-Z\s]+)`),
				regexp.MustCompile(`(?i)register\s+(?:new\s+)?(?:payee|beneficiary)`),
			},
			Keywords: map[string]float64{
				"payee":       1.0,
				"beneficiary": 0.9,
				"add":         0.8,
				"save":        0.7,
				"new":         0.6,
				"register":    0.6,
			},
			IntentType: model.IntentAddBeneficiary,
		},
		{
			Name: "apply_loan",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)apply\s+(?:for\s+)?(?:a\s+)?(personal|home|car|business)\s+loan`),
				regexp.MustCompile(`(?i)need\s+(?:a\s+)?loan`),
				regexp.MustCompile(`(?i)loan\s+application`),
			},
			Keywords: map[string]float64{
				"loan":     1.0,
				"apply":    0.9,
				"borrow":   0.8,
				"credit":   0.7,
				"emi":      0.6,
				"personal": 0.5,
				"home":     0.5,
			},
			IntentType: model.IntentApplyLoan,
		},
		{
			Name: "credit_score",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)credit\s+score`),
				regexp.MustCompile(`(?i)cibil\s+score`),
				regexp.MustCompile(`(?i)credit\s+rating`),
				regexp.MustCompile(`(?i)what\s+is\s+my\s+credit\s+score`),
			},
			Keywords: map[string]float64{
				"credit": 1.0,
				"score":  0.9,
				"cibil":  0.8,
				"rating": 0.7,
			},
			IntentType: model.IntentCreditScore,
		},
		{
			Name: "conversational",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)^(hello|hi|hey|greetings|good\s+(morning|afternoon|evening))`),
				regexp.MustCompile(`(?i)^(how\s+are\s+you|how\s+do\s+you\s+do|what's\s+up|how's\s+it\s+going)`),
				regexp.MustCompile(`(?i)(what\s+(all|can|do)\s+(you|can\s+you)\s+(do|support|help|perform|operations))`),
				regexp.MustCompile(`(?i)(what\s+(are\s+)?(your\s+)?(capabilities|features|services|functions))`),
				regexp.MustCompile(`(?i)(tell\s+me\s+(about\s+)?(what\s+you\s+can\s+do|your\s+capabilities))`),
				regexp.MustCompile(`(?i)^(thanks|thank\s+you|thank\s+you\s+very\s+much)`),
				regexp.MustCompile(`(?i)^(bye|goodbye|see\s+you|tata)`),
			},
			Keywords: map[string]float64{
				"hello":        1.0,
				"hi":           1.0,
				"hey":          0.9,
				"how are you":  0.95,
				"what can you": 0.9,
				"capabilities": 0.9,
				"operations":   0.8,
				"support":      0.8,
				"help":         0.7,
				"thanks":       0.9,
				"thank you":    0.9,
				"bye":          0.9,
			},
			IntentType: model.IntentConversational,
		},
	}
}

// parseWithRules uses rule-based parsing with pattern matching and weighted keywords
func (ip *IntentParser) parseWithRules(userInput string) (*model.Intent, error) {
	input := strings.ToLower(strings.TrimSpace(userInput))

	// First check for conversational queries (greetings, capability questions, etc.)
	// These should be detected before other intents to avoid false matches
	conversationalPattern := ip.patterns[len(ip.patterns)-1] // Last pattern is conversational
	if conversationalPattern.IntentType == model.IntentConversational {
		confidence := ip.calculateConfidence(input, &conversationalPattern)
		if confidence >= 0.5 {
			return &model.Intent{
				Type:         model.IntentConversational,
				Confidence:   confidence,
				Entities:     make(map[string]interface{}),
				OriginalText: userInput,
			}, nil
		}
	}

	// Find best matching intent for banking operations
	var bestIntent *IntentPattern
	highestConfidence := 0.0

	for i := range ip.patterns {
		pattern := &ip.patterns[i]
		// Skip conversational pattern in main loop (already checked)
		if pattern.IntentType == model.IntentConversational {
			continue
		}
		confidence := ip.calculateConfidence(input, pattern)

		if confidence > highestConfidence {
			highestConfidence = confidence
			bestIntent = pattern
		}
	}

	// Extract entities based on best intent
	entities := ip.extractEntities(userInput, bestIntent)

	// If no good match, default to conversational (for friendly fallback) if it's a short query
	if bestIntent == nil || highestConfidence < 0.3 {
		// If it's a very short query (likely a greeting or question), treat as conversational
		if len(strings.Fields(input)) <= 5 {
			return &model.Intent{
				Type:         model.IntentConversational,
				Confidence:   0.6,
				Entities:     entities,
				OriginalText: userInput,
			}, nil
		}
		return &model.Intent{
			Type:         model.IntentUnknown,
			Confidence:   highestConfidence,
			Entities:     entities,
			OriginalText: userInput,
		}, nil
	}

	return &model.Intent{
		Type:         bestIntent.IntentType,
		Confidence:   highestConfidence,
		Entities:     entities,
		OriginalText: userInput,
	}, nil
}

// calculateConfidence calculates confidence score for an intent pattern
func (ip *IntentParser) calculateConfidence(message string, pattern *IntentPattern) float64 {
	confidence := 0.0
	words := strings.Fields(message)

	// Check regex patterns (higher weight for exact matches)
	for _, regexPattern := range pattern.Patterns {
		if regexPattern.MatchString(message) {
			confidence += 0.5
		}
	}

	// Check weighted keywords (both single words and multi-word phrases)
	keywordMatches := 0
	for keyword, weight := range pattern.Keywords {
		// Check if keyword is a multi-word phrase
		if strings.Contains(keyword, " ") {
			if strings.Contains(message, keyword) {
				confidence += weight
				keywordMatches++
			}
		} else {
			// Single word keyword
			for _, word := range words {
				if word == keyword {
					confidence += weight
					keywordMatches++
					break // Only count once per keyword
				}
			}
		}
	}

	// Normalize confidence (cap at 1.0)
	if confidence > 1.0 {
		confidence = 1.0
	}

	// Boost confidence if multiple keywords match
	if keywordMatches > 1 {
		confidence = confidence * 1.1
		if confidence > 1.0 {
			confidence = 1.0
		}
	}

	return confidence
}

// extractEntities extracts entities from user input based on intent
func (ip *IntentParser) extractEntities(message string, pattern *IntentPattern) map[string]interface{} {
	entities := make(map[string]interface{})
	input := strings.ToLower(message)

	// Extract amount (common for transfers)
	amountRegex := regexp.MustCompile(`(?i)(?:rs\.?|₹|rupees?)?\s*(\d+(?:,\d{3})*(?:\.\d{2})?)\s*(?:rs|rupees|inr)?`)
	if matches := amountRegex.FindStringSubmatch(message); len(matches) > 1 {
		amountStr := strings.ReplaceAll(matches[1], ",", "")
		entities["amount"] = amountStr
	}

	// Extract account number
	accountRegex := regexp.MustCompile(`(?i)(?:account|acc|ac)\s*(?:no|number|#)?\s*:?\s*([\dX]{4,})`)
	if matches := accountRegex.FindStringSubmatch(message); len(matches) > 1 {
		entities["to_account"] = matches[1]
	}

	// Extract IFSC code
	ifscRegex := regexp.MustCompile(`(?i)ifsc\s*:?\s*([A-Z]{4}0[A-Z0-9]{6})`)
	if matches := ifscRegex.FindStringSubmatch(message); len(matches) > 1 {
		entities["ifsc"] = matches[1]
	}

	// Extract UPI ID
	upiRegex := regexp.MustCompile(`(?i)([a-zA-Z0-9._-]+@[a-zA-Z0-9]+)`)
	if matches := upiRegex.FindStringSubmatch(message); len(matches) > 1 {
		entities["upi_id"] = matches[1]
	}

	// Extract payee/beneficiary name
	if pattern != nil && (pattern.IntentType == model.IntentAddBeneficiary || pattern.IntentType == model.IntentTransferNEFT || pattern.IntentType == model.IntentTransferRTGS || pattern.IntentType == model.IntentTransferIMPS) {
		nameRegex := regexp.MustCompile(`(?i)(?:to|for|payee|beneficiary|named)\s+([a-zA-Z\s]{2,})`)
		if matches := nameRegex.FindStringSubmatch(message); len(matches) > 1 {
			entities["payee_name"] = strings.TrimSpace(matches[1])
		}
	}

	// Extract loan type
	if pattern != nil && pattern.IntentType == model.IntentApplyLoan {
		loanTypes := []string{"personal", "home", "car", "business"}
		for _, loanType := range loanTypes {
			if strings.Contains(input, loanType) {
				entities["loan_type"] = loanType
				break
			}
		}
	}

	// Extract transfer method if not already specified
	if pattern != nil && (pattern.IntentType == model.IntentTransferNEFT || pattern.IntentType == model.IntentTransferRTGS || pattern.IntentType == model.IntentTransferIMPS || pattern.IntentType == model.IntentTransferUPI) {
		methods := map[string]string{
			"upi":  "UPI",
			"imps": "IMPS",
			"neft": "NEFT",
			"rtgs": "RTGS",
		}
		for method, value := range methods {
			if strings.Contains(input, method) {
				entities["method"] = value
				break
			}
		}
		// Default to UPI if no method specified but it's a transfer
		if _, exists := entities["method"]; !exists {
			entities["method"] = "UPI"
		}
	}

	return entities
}
