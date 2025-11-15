package model

import "time"

// OrchestrationPlan represents a plan for multi-agent orchestration
type OrchestrationPlan struct {
	PlanID          string                 `json:"plan_id"`
	UserID          string                 `json:"user_id"`
	Intent          Intent                 `json:"intent"`
	RequiredAgents  []AgentRequirement     `json:"required_agents"`
	ExecutionOrder  []string               `json:"execution_order"` // Agent IDs in execution order
	ParallelAgents  [][]string             `json:"parallel_agents"` // Agents that can run in parallel
	Context         map[string]interface{} `json:"context"`
	CreatedAt       time.Time              `json:"created_at"`
}

// AgentRequirement represents a requirement for a specific agent
type AgentRequirement struct {
	AgentType       string                 `json:"agent_type"`
	Priority        int                    `json:"priority"` // Higher = more important
	Required        bool                   `json:"required"` // If false, can skip if unavailable
	InputContext    map[string]interface{} `json:"input_context"`
	ExpectedOutput  string                 `json:"expected_output,omitempty"`
}

// ExecutionStep represents a step in the orchestration execution
type ExecutionStep struct {
	StepID      string                 `json:"step_id"`
	PlanID      string                 `json:"plan_id"`
	AgentID     string                 `json:"agent_id"`
	AgentType   string                 `json:"agent_type"`
	Status      string                 `json:"status"` // PENDING, RUNNING, COMPLETED, FAILED
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// OrchestrationResult represents the final result of orchestration
type OrchestrationResult struct {
	PlanID          string                 `json:"plan_id"`
	Status          string                 `json:"status"` // SUCCESS, PARTIAL, FAILED
	FinalResponse   MergedResponse         `json:"final_response"`
	Steps           []ExecutionStep        `json:"steps"`
	TotalDuration   time.Duration          `json:"total_duration"`
	CompletedAt     time.Time              `json:"completed_at"`
}

