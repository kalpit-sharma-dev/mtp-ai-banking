package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	log.Info().
		Str("task_id", task.TaskID).
		Str("agent_type", string(agent.Type)).
		Interface("result", result).
		Msg("Agent response received")

	// For transfers, if Guardrail approves, chain to Banking agent to execute
	if task.Intent == "TRANSFER_NEFT" || task.Intent == "TRANSFER_RTGS" || 
	   task.Intent == "TRANSFER_IMPS" || task.Intent == "TRANSFER_UPI" {
		if agent.Type == model.AgentTypeGuardrail {
			// Check if guardrail approved - check both result["status"] and result["all_passed"]
			statusApproved := false
			if status, ok := result["status"].(string); ok && status == "APPROVED" {
				statusApproved = true
			}
			// Also check all_passed field from guardrail checks
			allPassed := false
			if passed, ok := result["all_passed"].(bool); ok && passed {
				allPassed = true
			}
			
			if statusApproved || allPassed {
				// Chain to Banking agent to execute the transfer
				bankingAgents, err := o.agentRegistry.FindAgentsByType(ctx, model.AgentTypeBanking)
				if err == nil && len(bankingAgents) > 0 {
					bankingAgent := bankingAgents[0]
					log.Info().
						Str("task_id", task.TaskID).
						Str("banking_agent", bankingAgent.AgentID).
						Bool("status_approved", statusApproved).
						Bool("all_passed", allPassed).
						Msg("✅ Guardrail approved, chaining to Banking agent")
					
					// Call Banking agent with same request
					bankingResult, bankingRiskScore, bankingExplanation, err := o.callAgent(ctx, bankingAgent, agentRequest)
					if err != nil {
						log.Error().
							Err(err).
							Str("task_id", task.TaskID).
							Msg("❌ Failed to call Banking agent after Guardrail approval")
						o.taskManager.UpdateTaskStatus(ctx, task.TaskID, model.TaskStatusFailed, nil, fmt.Sprintf("Guardrail approved but Banking agent failed: %s", err.Error()))
						return
					}
					
					// Use Banking agent result (which has transaction_id, etc.)
					result = bankingResult
					riskScore = bankingRiskScore
					explanation = bankingExplanation
					
					log.Info().
						Str("task_id", task.TaskID).
						Interface("banking_result", bankingResult).
						Str("explanation", bankingExplanation).
						Msg("✅ Banking agent executed transfer successfully")
				} else {
					log.Error().
						Str("task_id", task.TaskID).
						Msg("❌ No Banking agent found to execute transfer")
				}
			} else {
				// Guardrail rejected, stop here
				log.Info().
					Str("task_id", task.TaskID).
					Interface("guardrail_result", result).
					Msg("❌ Guardrail rejected transfer, not chaining to Banking agent")
			}
		}
	}

	// Update task with result
	if err := o.taskManager.UpdateTaskResult(ctx, task.TaskID, result, riskScore, explanation); err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to update task result")
	}
}

// callAgent calls the agent's REST endpoint
func (o *Orchestrator) callAgent(ctx context.Context, agent *model.Agent, request map[string]interface{}) (map[string]interface{}, float64, string, error) {
	// Make actual HTTP call to agent endpoint
	if agent.Endpoint == "" {
		log.Warn().Str("agent_id", agent.AgentID).Msg("Agent endpoint is empty, using mock")
		return o.mockAgentByType(agent.Type, request)
	}

	// Prepare request payload for agent
	agentRequest := map[string]interface{}{
		"agent_id":      agent.AgentID,
		"request_id":    fmt.Sprintf("req_%d", time.Now().UnixNano()),
		"task":          request["task"],
		"input_context": request["input_context"],
		"session_id":    request["session_id"],
	}

	// Marshal request body
	body, err := json.Marshal(agentRequest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal agent request")
		return o.mockAgentByType(agent.Type, request)
	}

	// Build agent URL
	agentURL := fmt.Sprintf("%s/api/v1/process", agent.Endpoint)

	log.Info().
		Str("agent_id", agent.AgentID).
		Str("agent_type", string(agent.Type)).
		Str("endpoint", agentURL).
		Msg("Calling agent endpoint")

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", agentURL, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Str("agent_id", agent.AgentID).Msg("Failed to create HTTP request")
		return o.mockAgentByType(agent.Type, request)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", "test-api-key") // Use same API key as agents expect

	// Make HTTP call
	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		log.Error().
			Err(err).
			Str("agent_id", agent.AgentID).
			Str("endpoint", agentURL).
			Msg("Failed to call agent, falling back to mock")
		return o.mockAgentByType(agent.Type, request)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Str("agent_id", agent.AgentID).Msg("Failed to read agent response")
		return o.mockAgentByType(agent.Type, request)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("agent_id", agent.AgentID).
			Str("response", string(respBody)).
			Msg("Agent returned error status, falling back to mock")
		return o.mockAgentByType(agent.Type, request)
	}

	// Parse agent response
	var agentResponse struct {
		AgentID     string                 `json:"agent_id"`
		AgentType   string                 `json:"agent_type"`
		Status      string                 `json:"status"`
		Result      map[string]interface{} `json:"result"`
		RiskScore   float64                `json:"risk_score"`
		Explanation string                 `json:"explanation"`
		Confidence  float64                `json:"confidence"`
		RequestID   string                 `json:"request_id"`
	}

	if err := json.Unmarshal(respBody, &agentResponse); err != nil {
		log.Error().Err(err).Str("agent_id", agent.AgentID).Str("response", string(respBody)).Msg("Failed to parse agent response")
		return o.mockAgentByType(agent.Type, request)
	}

	log.Info().
		Str("agent_id", agent.AgentID).
		Str("status", agentResponse.Status).
		Float64("risk_score", agentResponse.RiskScore).
		Msg("Successfully called agent")

	// Return agent response
	return agentResponse.Result, agentResponse.RiskScore, agentResponse.Explanation, nil
}

// mockAgentByType returns mock response based on agent type (fallback)
func (o *Orchestrator) mockAgentByType(agentType model.AgentType, request map[string]interface{}) (map[string]interface{}, float64, string, error) {
	switch agentType {
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
