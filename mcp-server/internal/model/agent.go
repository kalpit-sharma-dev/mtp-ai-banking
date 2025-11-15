package model

import (
	"time"
)

// AgentType represents the type of agent
type AgentType string

const (
	AgentTypeBanking    AgentType = "BANKING"
	AgentTypeFraud      AgentType = "FRAUD"
	AgentTypeGuardrail  AgentType = "GUARDRAIL"
	AgentTypeClearance  AgentType = "CLEARANCE"
	AgentTypeScoring    AgentType = "SCORING"
	AgentTypePayment    AgentType = "PAYMENT"
	AgentTypeTrade      AgentType = "TRADE"
	AgentTypeAuth       AgentType = "AUTH"
)

// AgentStatus represents the health status of an agent
type AgentStatus string

const (
	AgentStatusHealthy   AgentStatus = "HEALTHY"
	AgentStatusUnhealthy AgentStatus = "UNHEALTHY"
	AgentStatusDegraded  AgentStatus = "DEGRADED"
)

// Agent represents a registered agent in the mesh
type Agent struct {
	AgentID      string                 `json:"agent_id" db:"agent_id"`
	Name         string                 `json:"name" db:"name"`
	Type         AgentType              `json:"type" db:"type"`
	Endpoint     string                 `json:"endpoint" db:"endpoint"`         // REST/gRPC endpoint
	GRPCEndpoint string                 `json:"grpc_endpoint,omitempty" db:"grpc_endpoint"`
	Status       AgentStatus            `json:"status" db:"status"`
	Capabilities []string               `json:"capabilities" db:"capabilities"` // What tasks it can handle
	Rules        map[string]interface{} `json:"rules" db:"rules"`               // Routing rules
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	HealthCheck  string                 `json:"health_check,omitempty" db:"health_check"`
	LastHealthAt time.Time              `json:"last_health_at" db:"last_health_at"`
	RegisteredAt time.Time              `json:"registered_at" db:"registered_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// AgentRegistrationRequest represents a request to register a new agent
type AgentRegistrationRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Type         string                 `json:"type" binding:"required"`
	Endpoint     string                 `json:"endpoint" binding:"required"`
	GRPCEndpoint string                 `json:"grpc_endpoint,omitempty"`
	Capabilities []string               `json:"capabilities" binding:"required"`
	Rules        map[string]interface{} `json:"rules,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	HealthCheck  string                 `json:"health_check,omitempty"`
}

// AgentResponse represents the agent registration response
type AgentResponse struct {
	AgentID      string    `json:"agent_id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	RegisteredAt time.Time `json:"registered_at"`
	Message      string    `json:"message"`
}

