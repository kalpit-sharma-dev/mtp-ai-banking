package service

import (
	"context"
	"fmt"
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
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(
	intentParser *IntentParser,
	contextEnricher *ContextEnricher,
	mcpClient *MCPClient,
	responseMerger *ResponseMerger,
) *Orchestrator {
	return &Orchestrator{
		intentParser:    intentParser,
		contextEnricher: contextEnricher,
		mcpClient:       mcpClient,
		responseMerger:  responseMerger,
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

