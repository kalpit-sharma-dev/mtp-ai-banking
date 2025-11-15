package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aibanking/mcp-server/internal/model"
	"github.com/rs/zerolog/log"
)

// Orchestrator coordinates task execution across agents
type Orchestrator struct {
	sessionManager *SessionManager
	taskManager    *TaskManager
	agentRegistry  *AgentRegistry
	contextRouter  *ContextRouter
	httpClient     *http.Client
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(
	sessionManager *SessionManager,
	taskManager *TaskManager,
	agentRegistry *AgentRegistry,
	contextRouter *ContextRouter,
) *Orchestrator {
	return &Orchestrator{
		sessionManager: sessionManager,
		taskManager:    taskManager,
		agentRegistry:  agentRegistry,
		contextRouter:  contextRouter,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessTask processes a task through the orchestration pipeline
func (o *Orchestrator) ProcessTask(ctx context.Context, req *model.TaskRequest) (*model.TaskResponse, error) {
	// Get or create session
	var session *model.Session
	var err error

	if req.SessionID != "" {
		session, err = o.sessionManager.GetSession(ctx, req.SessionID)
		if err != nil {
			// Create new session if not found
			sessionReq := &model.SessionRequest{
				UserID:  req.UserID,
				Channel: req.Channel,
				Context: req.Context,
			}
			session, err = o.sessionManager.CreateSession(ctx, sessionReq)
			if err != nil {
				return nil, fmt.Errorf("failed to create session: %w", err)
			}
		}
	} else {
		// Create new session
		sessionReq := &model.SessionRequest{
			UserID:  req.UserID,
			Channel: req.Channel,
			Context: req.Context,
		}
		session, err = o.sessionManager.CreateSession(ctx, sessionReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	}

	// Create task
	task, err := o.taskManager.CreateTask(ctx, req, session.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Add task to session
	if err := o.sessionManager.AddTaskToSession(ctx, session.SessionID, task.TaskID); err != nil {
		log.Warn().Err(err).Msg("Failed to add task to session")
	}

	// Route task to appropriate agent
	decision, err := o.contextRouter.RouteTask(ctx, task, session)
	if err != nil {
		o.taskManager.UpdateTaskStatus(ctx, task.TaskID, model.TaskStatusFailed, nil, err.Error())
		return nil, fmt.Errorf("failed to route task: %w", err)
	}

	// If no agent found, mark as failed
	if decision.SelectedAgentID == "" {
		o.taskManager.UpdateTaskStatus(ctx, task.TaskID, model.TaskStatusFailed, nil, "No agent available for routing")
		return nil, fmt.Errorf("no agent available for task routing")
	}

	// Update task with selected agent
	if err := o.taskManager.UpdateTaskAgent(ctx, task.TaskID, decision.SelectedAgentID); err != nil {
		return nil, fmt.Errorf("failed to update task agent: %w", err)
	}

	// Execute task asynchronously
	go o.executeTask(context.Background(), task, decision)

	return &model.TaskResponse{
		TaskID:    task.TaskID,
		SessionID: session.SessionID,
		Status:    string(task.Status),
		Message:   "Task submitted successfully",
		CreatedAt: task.CreatedAt,
	}, nil
}

// executeTask executes the task by calling the appropriate agent
func (o *Orchestrator) executeTask(ctx context.Context, task *model.Task, decision *model.RoutingDecision) {
	agent, err := o.agentRegistry.GetAgent(ctx, decision.SelectedAgentID)
	if err != nil {
		o.taskManager.UpdateTaskStatus(ctx, task.TaskID, model.TaskStatusFailed, nil, fmt.Sprintf("Agent not found: %s", err.Error()))
		return
	}

	// Prepare agent request payload
	agentRequest := map[string]interface{}{
		"agent_id": agent.AgentID,
		"task":     task.Intent,
		"input_context": map[string]interface{}{
			"user_id":    task.UserID,
			"session_id": task.SessionID,
			"channel":    task.Channel,
			"intent":     task.Intent,
			"data":       task.Data,
			"context":    task.Context,
		},
		"session_id": task.SessionID,
	}

	// Call agent endpoint
	result, riskScore, explanation, err := o.callAgent(ctx, agent, agentRequest)
	if err != nil {
		o.taskManager.UpdateTaskStatus(ctx, task.TaskID, model.TaskStatusFailed, nil, err.Error())
		return
	}

	// Update task with result
	if err := o.taskManager.UpdateTaskResult(ctx, task.TaskID, result, riskScore, explanation); err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to update task result")
	}
}

// callAgent calls the agent's REST endpoint
func (o *Orchestrator) callAgent(ctx context.Context, agent *model.Agent, request map[string]interface{}) (map[string]interface{}, float64, string, error) {
	// For now, use mock responses based on agent type
	// In production, this would make actual HTTP/gRPC calls

	switch agent.Type {
	case model.AgentTypeBanking:
		return o.mockBankingAgent(request)
	case model.AgentTypeFraud:
		return o.mockFraudAgent(request)
	case model.AgentTypeGuardrail:
		return o.mockGuardrailAgent(request)
	case model.AgentTypeClearance:
		return o.mockClearanceAgent(request)
	case model.AgentTypeScoring:
		return o.mockScoringAgent(request)
	default:
		return map[string]interface{}{"status": "processed"}, 0.1, "Task processed by default agent", nil
	}
}

// Mock agent implementations (to be replaced with actual HTTP calls)
func (o *Orchestrator) mockBankingAgent(request map[string]interface{}) (map[string]interface{}, float64, string, error) {
	inputCtx := request["input_context"].(map[string]interface{})
	intent := inputCtx["intent"].(string)

	result := map[string]interface{}{
		"status":  "APPROVED",
		"message": "Transaction processed successfully",
	}
	fmt.Println("####################", intent)
	if intent == "CHECK_BALANCE" {
		result["balance"] = 50000.0
		result["currency"] = "INR"
	}

	return result, 0.1, "Transaction is within user limits and behavior pattern is normal.", nil
}

func (o *Orchestrator) mockFraudAgent(request map[string]interface{}) (map[string]interface{}, float64, string, error) {
	inputCtx := request["input_context"].(map[string]interface{})
	data := inputCtx["data"].(map[string]interface{})

	amount, _ := data["amount"].(float64)
	riskScore := 0.3

	if amount > 100000 {
		riskScore = 0.7
		return map[string]interface{}{
			"status":      "REJECTED",
			"fraud_score": riskScore,
			"reason":      "High amount transaction flagged",
		}, riskScore, "Transaction flagged for manual review due to high amount.", nil
	}

	return map[string]interface{}{
		"status":      "APPROVED",
		"fraud_score": riskScore,
	}, riskScore, "No fraud patterns detected.", nil
}

func (o *Orchestrator) mockGuardrailAgent(request map[string]interface{}) (map[string]interface{}, float64, string, error) {
	return map[string]interface{}{
		"status":          "APPROVED",
		"guardrail_check": "PASSED",
		"rules_validated": []string{"daily_limit", "velocity_check", "beneficiary_age"},
	}, 0.15, "All guardrail rules passed.", nil
}

func (o *Orchestrator) mockClearanceAgent(request map[string]interface{}) (map[string]interface{}, float64, string, error) {
	return map[string]interface{}{
		"status":          "APPROVED",
		"clearance_level": "AUTO",
		"loan_amount":     100000,
		"interest_rate":   8.5,
	}, 0.2, "Loan application approved automatically based on credit score.", nil
}

func (o *Orchestrator) mockScoringAgent(request map[string]interface{}) (map[string]interface{}, float64, string, error) {
	return map[string]interface{}{
		"credit_score":   750,
		"risk_category":  "LOW",
		"recommendation": "APPROVE",
	}, 0.1, "Credit score calculated based on user profile and history.", nil
}
