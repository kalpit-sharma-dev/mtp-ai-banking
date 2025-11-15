package model

import (
	"time"
)

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "PENDING"
	TaskStatusProcessing TaskStatus = "PROCESSING"
	TaskStatusCompleted  TaskStatus = "COMPLETED"
	TaskStatusFailed     TaskStatus = "FAILED"
	TaskStatusRejected   TaskStatus = "REJECTED"
)

// Task represents a banking task submitted to the MCP server
type Task struct {
	TaskID      string                 `json:"task_id" db:"task_id"`
	SessionID   string                 `json:"session_id" db:"session_id"`
	UserID      string                 `json:"user_id" db:"user_id"`
	Channel     string                 `json:"channel" db:"channel"` // MB, NB, Trade, etc.
	Intent      string                 `json:"intent" db:"intent"`   // TRANSFER_NEFT, CHECK_BALANCE, etc.
	Status      TaskStatus             `json:"status" db:"status"`
	Data        map[string]interface{} `json:"data" db:"data"`
	Context     map[string]interface{} `json:"context" db:"context"`
	AgentID     string                 `json:"agent_id,omitempty" db:"agent_id"`
	Result      map[string]interface{} `json:"result,omitempty" db:"result"`
	Error       string                 `json:"error,omitempty" db:"error"`
	RiskScore   float64                `json:"risk_score,omitempty" db:"risk_score"`
	Explanation string                 `json:"explanation,omitempty" db:"explanation"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
}

// TaskRequest represents the incoming task submission request
type TaskRequest struct {
	SessionID string                 `json:"session_id"`
	UserID    string                 `json:"user_id" binding:"required"`
	Channel   string                 `json:"channel" binding:"required"`
	Intent    string                 `json:"intent" binding:"required"`
	Data      map[string]interface{} `json:"data" binding:"required"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// TaskResponse represents the response after task submission
type TaskResponse struct {
	TaskID    string    `json:"task_id"`
	SessionID string    `json:"session_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskResultResponse represents the result of a completed task
type TaskResultResponse struct {
	TaskID      string                 `json:"task_id"`
	Status      string                 `json:"status"`
	Result      map[string]interface{} `json:"result,omitempty"`
	RiskScore   float64                `json:"risk_score,omitempty"`
	Explanation string                 `json:"explanation,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

