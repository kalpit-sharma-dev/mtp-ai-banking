package service

import (
	"context"
	"fmt"

	"github.com/aibanking/mcp-server/internal/model"
	"github.com/rs/zerolog/log"
)

// ContextRouter determines which agent should handle a task based on context
type ContextRouter struct {
	agentRegistry *AgentRegistry
	ruleEngine   *RuleEngine
}

// NewContextRouter creates a new context router instance
func NewContextRouter(agentRegistry *AgentRegistry, ruleEngine *RuleEngine) *ContextRouter {
	return &ContextRouter{
		agentRegistry: agentRegistry,
		ruleEngine:    ruleEngine,
	}
}

// RouteTask determines the appropriate agent for a task based on context
func (cr *ContextRouter) RouteTask(ctx context.Context, task *model.Task, session *model.Session) (*model.RoutingDecision, error) {
	// Build enriched context
	enrichedContext := cr.buildContext(task, session)

	// Apply routing rules
	decision, err := cr.ruleEngine.EvaluateRoutingRules(ctx, enrichedContext, task)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate routing rules: %w", err)
	}

	// If no agent selected by rules, use intent-based routing
	if decision.SelectedAgentID == "" {
		decision = cr.routeByIntent(ctx, task, enrichedContext)
	}

	log.Info().
		Str("task_id", task.TaskID).
		Str("intent", task.Intent).
		Str("selected_agent", decision.SelectedAgentID).
		Float64("confidence", decision.Confidence).
		Msg("Task routed to agent")

	return decision, nil
}

// buildContext enriches context with session and task data
func (cr *ContextRouter) buildContext(task *model.Task, session *model.Session) *model.Context {
	ctx := &model.Context{
		UserID:          task.UserID,
		SessionID:       task.SessionID,
		Channel:         task.Channel,
		Intent:          task.Intent,
		UserProfile:     make(map[string]interface{}),
		TransactionData: task.Data,
		DeviceInfo:      make(map[string]interface{}),
		HistoricalData:  make(map[string]interface{}),
		Rules:           make(map[string]interface{}),
		Metadata:        make(map[string]interface{}),
	}

	// Merge session context
	if session != nil {
		for k, v := range session.Context {
			ctx.Metadata[k] = v
		}
	}

	// Extract user profile from context if available
	if userProfile, ok := task.Context["user_profile"].(map[string]interface{}); ok {
		ctx.UserProfile = userProfile
	}

	// Extract device info if available
	if deviceInfo, ok := task.Context["device_info"].(map[string]interface{}); ok {
		ctx.DeviceInfo = deviceInfo
	}

	// Determine risk level based on transaction amount
	if amount, ok := task.Data["amount"].(float64); ok {
		if amount > 100000 {
			ctx.RiskLevel = "HIGH"
		} else if amount > 50000 {
			ctx.RiskLevel = "MEDIUM"
		} else {
			ctx.RiskLevel = "LOW"
		}
	} else {
		ctx.RiskLevel = "LOW"
	}

	return ctx
}

// routeByIntent routes task based on intent when rules don't match
func (cr *ContextRouter) routeByIntent(ctx context.Context, task *model.Task, enrichedContext *model.Context) *model.RoutingDecision {
	var agentType model.AgentType
	var reason string

	switch task.Intent {
	case "TRANSFER_NEFT", "TRANSFER_RTGS", "TRANSFER_IMPS", "TRANSFER_UPI":
		// First check guardrail, then fraud, then banking
		if cr.shouldRouteToGuardrail(enrichedContext) {
			agentType = model.AgentTypeGuardrail
			reason = "Transaction requires guardrail validation"
		} else if cr.shouldRouteToFraud(enrichedContext) {
			agentType = model.AgentTypeFraud
			reason = "Transaction flagged for fraud check"
		} else {
			agentType = model.AgentTypeBanking
			reason = "Standard banking transaction"
		}

	case "CHECK_BALANCE", "GET_STATEMENT", "VIEW_ACCOUNT":
		agentType = model.AgentTypeBanking
		reason = "Account inquiry operation"

	case "ADD_BENEFICIARY", "MANAGE_BENEFICIARY":
		agentType = model.AgentTypeGuardrail
		reason = "Beneficiary management requires validation"

	case "APPLY_LOAN", "LOAN_APPROVAL":
		agentType = model.AgentTypeClearance
		reason = "Loan application requires clearance"

	case "CREDIT_SCORE", "RISK_ASSESSMENT":
		agentType = model.AgentTypeScoring
		reason = "Credit/risk scoring operation"

	default:
		agentType = model.AgentTypeBanking
		reason = "Default routing to banking agent"
	}

	// Find available agent of this type
	agents, err := cr.agentRegistry.FindAgentsByType(ctx, agentType)
	if err != nil || len(agents) == 0 {
		log.Warn().
			Str("agent_type", string(agentType)).
			Msg("No agents found for type, using banking agent as fallback")
		agents, _ = cr.agentRegistry.FindAgentsByType(ctx, model.AgentTypeBanking)
	}

	if len(agents) == 0 {
		return &model.RoutingDecision{
			SelectedAgentID: "",
			AgentType:       string(agentType),
			Confidence:      0.0,
			Reason:          "No agents available",
			Context:         enrichedContext,
		}
	}

	// Select first available agent (can be enhanced with load balancing)
	selectedAgent := agents[0]

	return &model.RoutingDecision{
		SelectedAgentID: selectedAgent.AgentID,
		AgentType:       string(agentType),
		Confidence:      0.8,
		Reason:          reason,
		Context:         enrichedContext,
	}
}

// shouldRouteToGuardrail determines if transaction needs guardrail check
func (cr *ContextRouter) shouldRouteToGuardrail(ctx *model.Context) bool {
	// Route to guardrail for high-risk transactions or new beneficiaries
	if ctx.RiskLevel == "HIGH" {
		return true
	}

	// Check if beneficiary is new
	if txnData, ok := ctx.TransactionData["to_account"].(string); ok && txnData != "" {
		// In real implementation, check beneficiary age from database
		// For now, assume new if not in metadata
		if _, exists := ctx.Metadata["beneficiary_age_days"]; !exists {
			return true
		}
	}

	return false
}

// shouldRouteToFraud determines if transaction needs fraud check
func (cr *ContextRouter) shouldRouteToFraud(ctx *model.Context) bool {
	// Route to fraud for high-risk or medium-risk with suspicious patterns
	if ctx.RiskLevel == "HIGH" {
		return true
	}

	// Check for suspicious patterns in metadata
	if suspicious, ok := ctx.Metadata["suspicious_pattern"].(bool); ok && suspicious {
		return true
	}

	return false
}

