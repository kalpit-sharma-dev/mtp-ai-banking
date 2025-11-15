package model

import "time"

// AgentRequest represents a request to an agent
type AgentRequest struct {
	AgentID     string                 `json:"agent_id"`
	Task        string                 `json:"task"` // Intent type
	InputContext map[string]interface{} `json:"input_context"`
	SessionID   string                 `json:"session_id"`
	RequestID   string                 `json:"request_id"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AgentResponse represents a response from an agent
type AgentResponse struct {
	AgentID     string                 `json:"agent_id"`
	AgentType   string                 `json:"agent_type"`
	Status      string                 `json:"status"` // APPROVED, REJECTED, PENDING
	Result      map[string]interface{} `json:"result"`
	RiskScore   float64                `json:"risk_score"`
	Explanation string                 `json:"explanation"`
	Confidence  float64                `json:"confidence"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id"`
}

// BankingTransaction represents a banking transaction
type BankingTransaction struct {
	TransactionID string    `json:"transaction_id"`
	Type          string    `json:"type"` // NEFT, RTGS, IMPS, UPI
	FromAccount   string    `json:"from_account"`
	ToAccount     string    `json:"to_account"`
	Amount        float64   `json:"amount"`
	IFSC          string    `json:"ifsc,omitempty"`
	Remarks       string    `json:"remarks,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// FraudCheckRequest represents a fraud check request
type FraudCheckRequest struct {
	UserID        string                 `json:"user_id"`
	Transaction   BankingTransaction     `json:"transaction"`
	DeviceInfo    map[string]interface{} `json:"device_info,omitempty"`
	LocationInfo  map[string]interface{} `json:"location_info,omitempty"`
	UserBehavior  map[string]interface{} `json:"user_behavior,omitempty"`
	History       []BankingTransaction   `json:"history,omitempty"`
}

// GuardrailCheckRequest represents a guardrail validation request
type GuardrailCheckRequest struct {
	UserID      string                 `json:"user_id"`
	Transaction BankingTransaction     `json:"transaction"`
	UserProfile map[string]interface{} `json:"user_profile,omitempty"`
	Rules       []string               `json:"rules,omitempty"` // Which rules to check
}

// ClearanceRequest represents a loan clearance request
type ClearanceRequest struct {
	UserID      string                 `json:"user_id"`
	LoanType    string                 `json:"loan_type"` // PERSONAL, HOME, AUTO, etc.
	Amount      float64                `json:"amount"`
	Tenure      int                    `json:"tenure"` // months
	UserProfile map[string]interface{} `json:"user_profile"`
	CreditScore int                    `json:"credit_score,omitempty"`
	Income      float64                `json:"income,omitempty"`
}

// ScoringRequest represents a scoring request
type ScoringRequest struct {
	UserID        string                 `json:"user_id"`
	ScoreType     string                 `json:"score_type"` // CREDIT, FRAUD, RISK
	Transaction   *BankingTransaction    `json:"transaction,omitempty"`
	UserProfile   map[string]interface{} `json:"user_profile"`
	History       []BankingTransaction   `json:"history,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
}

