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
	ragService       *RAGService
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(
	intentParser *IntentParser,
	contextEnricher *ContextEnricher,
	mcpClient *MCPClient,
	responseMerger *ResponseMerger,
	llmService *LLMService,
	sessionService *SessionService,
	ragService *RAGService,
) *Orchestrator {
	return &Orchestrator{
		intentParser:    intentParser,
		contextEnricher: contextEnricher,
		mcpClient:       mcpClient,
		responseMerger:  responseMerger,
		llmService:      llmService,
		sessionService:  sessionService,
		ragService:      ragService,
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
		// Before rejecting, check if RAG has relevant context that might help
		// This allows RAG to handle queries that don't match standard intents
		if o.ragService != nil {
			// Try to retrieve relevant context
			relevantDocs, err := o.ragService.RetrieveRelevantContext(ctx, req.UserID, req.Input, 3)
			if err == nil && len(relevantDocs) > 0 {
				// If we have relevant context, treat as conversational query
				// This allows RAG to answer questions about past transactions/conversations
				log.Info().
					Str("user_id", req.UserID).
					Int("relevant_docs", len(relevantDocs)).
					Msg("Unknown intent but found relevant RAG context, treating as conversational")
				
				// Store user input in RAG
				o.ragService.StoreConversation(ctx, req.UserID, req.SessionID, "user", req.Input)
				
				// Handle as conversational query with RAG context
				return o.handleConversationalQuery(ctx, req)
			}
		}
		
		// Return a helpful error response if no relevant context found
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

	// Step 2: Store user input in RAG for context awareness
	if o.ragService != nil {
		o.ragService.StoreConversation(ctx, req.UserID, req.SessionID, "user", req.Input)
	}

	// Step 3: Enrich context with user history and behavior
	enrichedContext, err := o.contextEnricher.EnrichContext(ctx, req.UserID, req.SessionID, req.Channel, *intent)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich context: %w", err)
	}

	// Store user profile and transactions in RAG
	if o.ragService != nil {
		o.ragService.StoreUserContext(ctx, req.UserID, &enrichedContext.UserProfile)
		for _, txn := range enrichedContext.TransactionHistory {
			o.ragService.StoreTransaction(ctx, req.UserID, &txn)
		}
	}

	log.Info().
		Str("risk_level", enrichedContext.RiskIndicators.OverallRisk).
		Msg("Context enriched")

	// Step 4: Determine if multi-agent coordination is needed
	_ = o.shouldUseMultiAgent(intent, enrichedContext) // Reserved for future multi-agent coordination

	// Step 5: Submit task to MCP server and get response
	agentResponse, err := o.mcpClient.SubmitTask(ctx, req, *intent, enrichedContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent response: %w", err)
	}

	log.Info().
		Str("agent_status", agentResponse.Status).
		Float64("risk_score", agentResponse.RiskScore).
		Msg("Received agent response")

	// Step 6: Store agent response in RAG for future context
	if o.ragService != nil {
		// Store conversation
		if agentResponse.Explanation != "" {
			o.ragService.StoreConversation(ctx, req.UserID, req.SessionID, "assistant", agentResponse.Explanation)
		}
		
		// Extract and store transaction if present in response
		if agentResponse.Result != nil {
			// Check if this is a transfer/transaction response
			if txnID, ok := agentResponse.Result["transaction_id"].(string); ok && txnID != "" {
				// Create transaction record from agent response
				amount, _ := agentResponse.Result["amount"].(float64)
				status, _ := agentResponse.Result["status"].(string)
				if status == "" {
					status = agentResponse.Status
				}
				
				// Determine transaction type from intent
				txnType := "TRANSFER"
				if intent.Type == model.IntentTransferNEFT {
					txnType = "NEFT"
				} else if intent.Type == model.IntentTransferRTGS {
					txnType = "RTGS"
				} else if intent.Type == model.IntentTransferIMPS {
					txnType = "IMPS"
				} else if intent.Type == model.IntentTransferUPI {
					txnType = "UPI"
				}
				
				// Store transaction in RAG
				txnRecord := &model.TransactionRecord{
					TransactionID: txnID,
					Type:          txnType,
					Amount:        amount,
					Timestamp:     agentResponse.Timestamp,
					Status:        status,
				}
				o.ragService.StoreTransaction(ctx, req.UserID, txnRecord)
				
				log.Info().
					Str("transaction_id", txnID).
					Str("user_id", req.UserID).
					Msg("Stored transaction in RAG from agent response")
			}
			
			// Check if this is a beneficiary addition response
			if beneficiaryID, ok := agentResponse.Result["beneficiary_id"].(string); ok && beneficiaryID != "" {
				log.Info().
					Str("beneficiary_id", beneficiaryID).
					Str("user_id", req.UserID).
					Msg("Beneficiary added successfully - should be stored in Banking Integrations DWH")
				// Note: Beneficiary is already stored in Banking Integrations DWH by the agent
				// The UI will refresh via the event dispatched in AIAssistant.jsx
			}
		}
	}

	// Step 7: If multi-agent, we would coordinate multiple agents here
	// For now, we'll use single agent response
	responses := []model.AgentResponse{*agentResponse}

	// Step 8: Merge responses (even if single, for consistency)
	mergedResponse, err := o.responseMerger.MergeResponses(responses)
	if err != nil {
		return nil, fmt.Errorf("failed to merge responses: %w", err)
	}

	// Step 9: Also check merged response for transaction_id (in case it's in FinalResult)
	if mergedResponse.FinalResult != nil {
		if txnID, ok := mergedResponse.FinalResult["transaction_id"].(string); ok && txnID != "" {
			// Check if we already stored this transaction
			amount, _ := mergedResponse.FinalResult["amount"].(float64)
			status, _ := mergedResponse.FinalResult["status"].(string)
			if status == "" {
				status = mergedResponse.Status
			}
			
			// Determine transaction type from intent
			txnType := "TRANSFER"
			if intent.Type == model.IntentTransferNEFT {
				txnType = "NEFT"
			} else if intent.Type == model.IntentTransferRTGS {
				txnType = "RTGS"
			} else if intent.Type == model.IntentTransferIMPS {
				txnType = "IMPS"
			} else if intent.Type == model.IntentTransferUPI {
				txnType = "UPI"
			}
			
			// Store transaction in RAG if not already stored
			if o.ragService != nil {
				txnRecord := &model.TransactionRecord{
					TransactionID: txnID,
					Type:          txnType,
					Amount:        amount,
					Timestamp:     time.Now(),
					Status:        status,
				}
				o.ragService.StoreTransaction(ctx, req.UserID, txnRecord)
				
				log.Info().
					Str("transaction_id", txnID).
					Str("user_id", req.UserID).
					Float64("amount", amount).
					Msg("Stored transaction in RAG from merged response")
			}
		}
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
		Msg("Handling conversational query with RAG")

	// Store user input in RAG
	if o.ragService != nil {
		o.ragService.StoreConversation(ctx, req.UserID, req.SessionID, "user", req.Input)
	}

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

	// Build base prompt with RAG context
	basePrompt := ""
	if o.ragService != nil {
		// Get user context summary
		contextSummary, err := o.ragService.GetUserContextSummary(ctx, req.UserID)
		if err == nil && contextSummary != "" {
			basePrompt = contextSummary + "\n\n"
		}

		// Build RAG-augmented prompt
		ragPrompt, err := o.ragService.BuildRAGPrompt(ctx, req.UserID, req.Input, basePrompt)
		if err == nil {
			basePrompt = ragPrompt
		}
	}

	// Call LLM service with conversation history and RAG context
	if o.llmService != nil {
		// If we have RAG context, we need to build a custom prompt
		var response string
		var err error

		// Build full prompt with RAG context
		if basePrompt != "" {
			// Add banking system prompt
			fullPrompt := `You are a secure and intelligent AI banking assistant integrated into a digital banking system.

Your role is to help users perform a wide range of banking tasks safely, efficiently, and clearly. Always ensure user intent is well-understood, confirm sensitive operations, and provide helpful, accurate guidance at every step.

**IMPORTANT: Always identify yourself as a banking assistant in your responses, especially when responding to greetings or questions about your capabilities.**

` + basePrompt

			// Add conversation history
			if len(conversationHistory) > 0 {
				fullPrompt += "\n\n**Recent Conversation History:**\n"
				for _, msg := range conversationHistory {
					role := "User"
					if msg["role"] == "assistant" || msg["role"] == "bot" {
						role = "Assistant"
					}
					fullPrompt += fmt.Sprintf("%s: %s\n", role, msg["content"])
				}
				fullPrompt += "\n"
			}
			fullPrompt += fmt.Sprintf("User: %s\nAssistant:", req.Input)
			
			// Call LLM with full RAG-augmented prompt
			response, err = o.llmService.CallLLM(ctx, fullPrompt)
		} else {
			// Fallback to regular history-based call
			response, err = o.llmService.CallLLMWithHistory(ctx, req.Input, conversationHistory)
		}

		if err != nil {
			log.Warn().Err(err).Msg("LLM service failed, using fallback response")
			response = o.getFallbackConversationalResponse(req.Input)
		}

		// Store assistant response in RAG
		if o.ragService != nil {
			o.ragService.StoreConversation(ctx, req.UserID, req.SessionID, "assistant", response)
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
	
	// Store fallback response in RAG
	if o.ragService != nil {
		o.ragService.StoreConversation(ctx, req.UserID, req.SessionID, "assistant", response)
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

