package model

import "time"

// IntentType represents the type of banking intent
type IntentType string

const (
	IntentTransferNEFT   IntentType = "TRANSFER_NEFT"
	IntentTransferRTGS   IntentType = "TRANSFER_RTGS"
	IntentTransferIMPS   IntentType = "TRANSFER_IMPS"
	IntentTransferUPI    IntentType = "TRANSFER_UPI"
	IntentCheckBalance   IntentType = "CHECK_BALANCE"
	IntentGetStatement   IntentType = "GET_STATEMENT"
	IntentAddBeneficiary IntentType = "ADD_BENEFICIARY"
	IntentApplyLoan      IntentType = "APPLY_LOAN"
	IntentCreditScore    IntentType = "CREDIT_SCORE"
	IntentUnknown        IntentType = "UNKNOWN"
)

// Intent represents a parsed user intent
type Intent struct {
	Type        IntentType              `json:"type"`
	Confidence  float64                 `json:"confidence"` // 0.0 to 1.0
	Entities    map[string]interface{}  `json:"entities"`    // Extracted entities (amount, account, etc.)
	OriginalText string                 `json:"original_text,omitempty"`
	Metadata    map[string]interface{}  `json:"metadata,omitempty"`
}

// UserRequest represents the incoming user request
type UserRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	Channel     string                 `json:"channel" binding:"required"` // MB, NB, etc.
	Input       string                 `json:"input"`                      // Natural language or structured
	InputType   string                 `json:"input_type"`                // "natural_language" or "structured"
	StructuredData map[string]interface{} `json:"structured_data,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
}

// OrchestrationRequest represents a request to the orchestrator
type OrchestrationRequest struct {
	UserID      string                 `json:"user_id"`
	Channel     string                 `json:"channel"`
	Intent      Intent                 `json:"intent"`
	Context     map[string]interface{} `json:"context"`
	SessionID   string                 `json:"session_id"`
	RequiresMultiAgent bool            `json:"requires_multi_agent"`
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
}

// MergedResponse represents the final merged response from multiple agents
type MergedResponse struct {
	Status      string                 `json:"status"` // APPROVED, REJECTED, PENDING, CONFLICT
	FinalResult map[string]interface{} `json:"final_result"`
	RiskScore   float64                `json:"risk_score"`
	Explanation string                 `json:"explanation"`
	AgentResponses []AgentResponse     `json:"agent_responses"`
	Conflicts   []Conflict             `json:"conflicts,omitempty"`
	ResolvedBy  string                 `json:"resolved_by,omitempty"` // Which agent/rule resolved conflicts
}

// Conflict represents a conflict between agent responses
type Conflict struct {
	Type        string                 `json:"type"` // "STATUS_MISMATCH", "RISK_SCORE_MISMATCH", etc.
	Description string                 `json:"description"`
	Agents      []string               `json:"agents"` // Agent IDs involved
	Values      map[string]interface{} `json:"values"` // Conflicting values
}

// EnrichedContext represents context enriched with user history and patterns
type EnrichedContext struct {
	UserID          string                 `json:"user_id"`
	SessionID       string                 `json:"session_id"`
	Channel         string                 `json:"channel"`
	Intent          Intent                 `json:"intent"`
	UserProfile     UserProfile            `json:"user_profile"`
	TransactionHistory []TransactionRecord `json:"transaction_history,omitempty"`
	RiskIndicators  RiskIndicators         `json:"risk_indicators"`
	BehaviorPattern BehaviorPattern        `json:"behavior_pattern"`
	Metadata      map[string]interface{}    `json:"metadata"`
}

// UserProfile represents user profile information
type UserProfile struct {
	AccountAge      int     `json:"account_age_days"`
	TotalBalance    float64 `json:"total_balance"`
	MonthlyIncome   float64 `json:"monthly_income,omitempty"`
	CreditScore     int     `json:"credit_score,omitempty"`
	KYCStatus       string  `json:"kyc_status"`
	AccountType     string  `json:"account_type"`
	TransactionCount int    `json:"transaction_count_30d"`
}

// TransactionRecord represents a historical transaction
type TransactionRecord struct {
	TransactionID string    `json:"transaction_id"`
	Type          string    `json:"type"`
	Amount        float64   `json:"amount"`
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"`
}

// RiskIndicators represents risk assessment indicators
type RiskIndicators struct {
	OverallRisk    string   `json:"overall_risk"` // LOW, MEDIUM, HIGH
	FraudRisk      float64  `json:"fraud_risk"`
	CreditRisk     float64  `json:"credit_risk"`
	VelocityRisk   float64  `json:"velocity_risk"`
	AmountRisk     float64  `json:"amount_risk"`
	DeviceRisk     float64  `json:"device_risk,omitempty"`
	LocationRisk   float64  `json:"location_risk,omitempty"`
}

// BehaviorPattern represents user behavior patterns
type BehaviorPattern struct {
	AverageAmount      float64   `json:"average_amount"`
	PeakHours          []int     `json:"peak_hours"` // Hours of day when user is most active
	CommonChannels     []string  `json:"common_channels"`
	FrequentBeneficiaries []string `json:"frequent_beneficiaries"`
	AnomalyDetected   bool      `json:"anomaly_detected"`
}

