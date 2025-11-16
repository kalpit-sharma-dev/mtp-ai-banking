package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aibanking/ai-skin-orchestrator/internal/model"
	"github.com/rs/zerolog/log"
)

// Orchestrator is the main AI Skin Orchestrator that coordinates all services
type Orchestrator struct {
	intentParser     *IntentParser
	contextEnricher  *ContextEnricher
	mcpClient        *MCPClient
	responseMerger   *ResponseMerger
	llmService       *LLMService
	sessionService   *SessionService
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(
	intentParser *IntentParser,
	contextEnricher *ContextEnricher,
	mcpClient *MCPClient,
	responseMerger *ResponseMerger,
	llmService *LLMService,
	sessionService *SessionService,
) *Orchestrator {
	return &Orchestrator{
		intentParser:    intentParser,
		contextEnricher: contextEnricher,
		mcpClient:       mcpClient,
		responseMerger:  responseMerger,
		llmService:      llmService,
		sessionService:  sessionService,
	}
}

// ProcessRequest processes a user request through the full orchestration pipeline
func (o *Orchestrator) ProcessRequest(ctx context.Context, req *model.UserRequest) (*model.MergedResponse, error) {
	startTime := time.Now()

	log.Info().
		Str("user_id", req.UserID).
		Str("channel", req.Channel).
		Str("input_type", req.InputType).
		Msg("Processing user request")

	// Step 1: Parse intent from user input
	intent, err := o.intentParser.ParseIntent(ctx, req.Input, req.InputType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse intent: %w", err)
	}

	if intent.Type == model.IntentUnknown {
		// Return a helpful error response instead of failing
		return &model.MergedResponse{
			Status: "REJECTED",
			FinalResult: map[string]interface{}{
				"error": "Could not understand your request. Please try rephrasing or use one of these: check balance, transfer money, view statement, add beneficiary.",
			},
			RiskScore:   0.5,
			Explanation: "I couldn't determine what you're asking for. Please try phrases like 'Check my balance', 'Transfer money', 'Show statement', or 'Add beneficiary'.",
			AgentResponses: []model.AgentResponse{},
		}, nil
	}

	log.Info().
		Str("intent", string(intent.Type)).
		Float64("confidence", intent.Confidence).
		Msg("Intent parsed")

	// Handle conversational queries directly with LLM (greetings, capability questions, etc.)
	if intent.Type == model.IntentConversational {
		return o.handleConversationalQuery(ctx, req)
	}

	// Step 2: Enrich context with user history and behavior
	enrichedContext, err := o.contextEnricher.EnrichContext(ctx, req.UserID, req.SessionID, req.Channel, *intent)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich context: %w", err)
	}

	log.Info().
		Str("risk_level", enrichedContext.RiskIndicators.OverallRisk).
		Msg("Context enriched")

	// Step 3: Determine if multi-agent coordination is needed
	_ = o.shouldUseMultiAgent(intent, enrichedContext) // Reserved for future multi-agent coordination

	// Step 4: Submit task to MCP server and get response
	agentResponse, err := o.mcpClient.SubmitTask(ctx, req, *intent, enrichedContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent response: %w", err)
	}

	log.Info().
		Str("agent_status", agentResponse.Status).
		Float64("risk_score", agentResponse.RiskScore).
		Msg("Received agent response")

	// Step 5: If multi-agent, we would coordinate multiple agents here
	// For now, we'll use single agent response
	responses := []model.AgentResponse{*agentResponse}

	// Step 6: Merge responses (even if single, for consistency)
	mergedResponse, err := o.responseMerger.MergeResponses(responses)
	if err != nil {
		return nil, fmt.Errorf("failed to merge responses: %w", err)
	}

	duration := time.Since(startTime)
	log.Info().
		Str("final_status", mergedResponse.Status).
		Dur("duration", duration).
		Msg("Request processed successfully")

	return mergedResponse, nil
}

// handleConversationalQuery handles greetings, capability questions, and other conversational queries
func (o *Orchestrator) handleConversationalQuery(ctx context.Context, req *model.UserRequest) (*model.MergedResponse, error) {
	log.Info().
		Str("user_id", req.UserID).
		Str("input", req.Input).
		Msg("Handling conversational query")

	// Get conversation history from session
	var conversationHistory []map[string]string
	if req.SessionID != "" && o.sessionService != nil {
		allHistory := o.sessionService.GetConversationHistory(req.SessionID)
		// Get last 10 messages for context
		start := 0
		if len(allHistory) > 10 {
			start = len(allHistory) - 10
		}
		conversationHistory = allHistory[start:]
	}

	// Call LLM service with conversation history
	if o.llmService != nil {
		response, err := o.llmService.CallLLMWithHistory(ctx, req.Input, conversationHistory)
		if err != nil {
			log.Warn().Err(err).Msg("LLM service failed, using fallback response")
			// Fallback response
			response = o.getFallbackConversationalResponse(req.Input)
		}

		return &model.MergedResponse{
			Status: "APPROVED",
			FinalResult: map[string]interface{}{
				"message": response,
				"type":    "conversational",
			},
			RiskScore:   0.0,
			Explanation: response,
			AgentResponses: []model.AgentResponse{},
		}, nil
	}

	// Fallback if LLM service not available
	response := o.getFallbackConversationalResponse(req.Input)
	return &model.MergedResponse{
		Status: "APPROVED",
		FinalResult: map[string]interface{}{
			"message": response,
			"type":    "conversational",
		},
		RiskScore:   0.0,
		Explanation: response,
		AgentResponses: []model.AgentResponse{},
	}, nil
}

// getFallbackConversationalResponse provides fallback responses for conversational queries
func (o *Orchestrator) getFallbackConversationalResponse(input string) string {
	inputLower := strings.ToLower(strings.TrimSpace(input))
	
	// Greetings
	if strings.HasPrefix(inputLower, "hello") || strings.HasPrefix(inputLower, "hi") || 
	   strings.HasPrefix(inputLower, "hey") || strings.HasPrefix(inputLower, "greetings") {
		return "Hello! I'm your AI banking assistant. How can I help you with your banking needs today?"
	}
	
	// How are you
	if strings.Contains(inputLower, "how are you") || strings.Contains(inputLower, "how do you do") {
		return "I'm doing great, thank you for asking! I'm here to help you with your banking operations. What would you like to do today?"
	}
	
	// Capability questions
	if strings.Contains(inputLower, "what can you") || strings.Contains(inputLower, "what do you") ||
	   strings.Contains(inputLower, "capabilities") || strings.Contains(inputLower, "operations") ||
	   strings.Contains(inputLower, "support") || strings.Contains(inputLower, "help") {
		return `I'm your AI banking assistant, and I can help you with the following operations:

• **Check Balance** - View your account balance
• **Fund Transfer** - Transfer money via NEFT, RTGS, IMPS, or UPI
• **View Statement** - Get your account statement and transaction history
• **Add Beneficiary** - Add a new payee for transfers
• **Create Fixed Deposit** - Open a fixed deposit account
• **Apply for Loan** - Apply for personal, home, or other loans
• **Credit Score** - Check your credit score

Just tell me what you'd like to do, and I'll help you with it!`
	}
	
	// Thanks
	if strings.HasPrefix(inputLower, "thank") || strings.HasPrefix(inputLower, "thanks") {
		return "You're welcome! Is there anything else I can help you with?"
	}
	
	// Goodbye
	if strings.HasPrefix(inputLower, "bye") || strings.HasPrefix(inputLower, "goodbye") {
		return "Goodbye! Have a great day. Feel free to come back if you need any banking assistance."
	}
	
	// Default
	return "I'm your AI banking assistant. I can help you with balance checks, fund transfers, statements, and more. What would you like to do?"
}

// shouldUseMultiAgent determines if multiple agents should be involved
func (o *Orchestrator) shouldUseMultiAgent(intent *model.Intent, context *model.EnrichedContext) bool {
	// Use multi-agent for high-risk transactions
	if context.RiskIndicators.OverallRisk == "HIGH" {
		return true
	}

	// Use multi-agent for large transfers
	if amount, ok := intent.Entities["amount"]; ok {
		if amountFloat, ok := amount.(float64); ok && amountFloat > 100000 {
			return true
		}
	}

	// Use multi-agent for loan applications
	if intent.Type == model.IntentApplyLoan {
		return true
	}

	return false
}

// ProcessMultiAgentRequest processes a request requiring multiple agents
func (o *Orchestrator) ProcessMultiAgentRequest(ctx context.Context, req *model.UserRequest, plan *model.OrchestrationPlan) (*model.OrchestrationResult, error) {
	// This would coordinate multiple agents in sequence or parallel
	// For now, this is a placeholder for future multi-agent coordination
	
	startTime := time.Now()
	var steps []model.ExecutionStep

	// Execute agents in order
	for _, agentType := range plan.ExecutionOrder {
		step := model.ExecutionStep{
			StepID:    fmt.Sprintf("step_%d", len(steps)+1),
			PlanID:    plan.PlanID,
			AgentType: agentType,
			Status:    "COMPLETED",
		}
		now := time.Now()
		step.StartedAt = &now
		step.CompletedAt = &now
		steps = append(steps, step)
	}

	duration := time.Since(startTime)

	return &model.OrchestrationResult{
		PlanID:        plan.PlanID,
		Status:        "SUCCESS",
		Steps:         steps,
		TotalDuration: duration,
		CompletedAt:   time.Now(),
	}, nil
}

