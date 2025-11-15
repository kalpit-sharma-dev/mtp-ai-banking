package model

// Context represents enriched context for task routing
type Context struct {
	UserID         string                 `json:"user_id"`
	SessionID      string                 `json:"session_id"`
	Channel        string                 `json:"channel"`
	Intent         string                 `json:"intent"`
	RiskLevel      string                 `json:"risk_level,omitempty"`      // LOW, MEDIUM, HIGH
	UserProfile    map[string]interface{} `json:"user_profile,omitempty"`   // Account age, transaction history, etc.
	TransactionData map[string]interface{} `json:"transaction_data,omitempty"` // Amount, beneficiary, etc.
	DeviceInfo     map[string]interface{} `json:"device_info,omitempty"`    // Device fingerprint, location
	HistoricalData map[string]interface{} `json:"historical_data,omitempty"` // Past transactions, patterns
	Rules          map[string]interface{} `json:"rules,omitempty"`          // Applicable business rules
	Metadata       map[string]interface{} `json:"metadata,omitempty"`       // Additional context
}

// RoutingDecision represents the decision made by the context router
type RoutingDecision struct {
	SelectedAgentID string                 `json:"selected_agent_id"`
	AgentType       string                 `json:"agent_type"`
	Confidence      float64                `json:"confidence"` // 0.0 to 1.0
	Reason          string                 `json:"reason"`
	AlternativeAgents []string             `json:"alternative_agents,omitempty"`
	Context         *Context               `json:"context"`
}

