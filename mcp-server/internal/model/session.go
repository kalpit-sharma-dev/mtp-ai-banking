package model

import (
	"time"
)

// Session represents a user session with context tracking
type Session struct {
	SessionID   string                 `json:"session_id" db:"session_id"`
	UserID      string                 `json:"user_id" db:"user_id"`
	Channel     string                 `json:"channel" db:"channel"`
	Context     map[string]interface{} `json:"context" db:"context"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	TaskHistory []string               `json:"task_history" db:"task_history"` // Array of task IDs
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	ExpiresAt   time.Time              `json:"expires_at" db:"expires_at"`
}

// SessionRequest represents a request to create or retrieve a session
type SessionRequest struct {
	UserID  string                 `json:"user_id" binding:"required"`
	Channel string                 `json:"channel" binding:"required"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// SessionResponse represents the session data response
type SessionResponse struct {
	SessionID   string                 `json:"session_id"`
	UserID      string                 `json:"user_id"`
	Channel     string                 `json:"channel"`
	Context     map[string]interface{} `json:"context"`
	TaskHistory []string               `json:"task_history"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

