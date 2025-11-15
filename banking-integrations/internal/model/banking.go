package model

import "time"

// Channel represents the banking channel
type Channel string

const (
	ChannelMB Channel = "MB" // Mobile Banking
	ChannelNB Channel = "NB" // Net Banking
	ChannelAPI Channel = "API" // API Banking
)

// TransactionType represents transaction type
type TransactionType string

const (
	TransactionTypeNEFT TransactionType = "NEFT"
	TransactionTypeRTGS TransactionType = "RTGS"
	TransactionTypeIMPS TransactionType = "IMPS"
	TransactionTypeUPI  TransactionType = "UPI"
	TransactionTypeDEBIT TransactionType = "DEBIT"
	TransactionTypeCREDIT TransactionType = "CREDIT"
)

// TransactionStatus represents transaction status
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
	TransactionStatusRejected  TransactionStatus = "REJECTED"
)

// Account represents a bank account
type Account struct {
	AccountID      string    `json:"account_id"`
	UserID         string    `json:"user_id"`
	AccountNumber  string    `json:"account_number"`
	AccountType    string    `json:"account_type"` // SAVINGS, CURRENT, etc.
	Balance        float64   `json:"balance"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"` // ACTIVE, INACTIVE, FROZEN
	KYCStatus      string    `json:"kyc_status"`
	CreatedAt      time.Time `json:"created_at"`
	LastUpdated    time.Time `json:"last_updated"`
}

// Transaction represents a banking transaction
type Transaction struct {
	TransactionID   string            `json:"transaction_id"`
	AccountID       string            `json:"account_id"`
	UserID          string            `json:"user_id"`
	Type            TransactionType   `json:"type"`
	Amount          float64           `json:"amount"`
	Currency        string            `json:"currency"`
	FromAccount     string            `json:"from_account,omitempty"`
	ToAccount       string            `json:"to_account,omitempty"`
	IFSC            string            `json:"ifsc,omitempty"`
	Status          TransactionStatus `json:"status"`
	Remarks         string            `json:"remarks,omitempty"`
	Channel         Channel           `json:"channel"`
	ReferenceNumber string            `json:"reference_number,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`
}

// Beneficiary represents a beneficiary
type Beneficiary struct {
	BeneficiaryID   string    `json:"beneficiary_id"`
	UserID          string    `json:"user_id"`
	AccountNumber   string    `json:"account_number"`
	IFSC            string    `json:"ifsc"`
	Name            string    `json:"name"`
	Nickname        string    `json:"nickname,omitempty"`
	AccountType     string    `json:"account_type"`
	Status          string    `json:"status"` // ACTIVE, INACTIVE
	AddedAt         time.Time `json:"added_at"`
	LastUsed        *time.Time `json:"last_used,omitempty"`
}

// StatementRequest represents a statement request
type StatementRequest struct {
	AccountID string    `json:"account_id"`
	UserID    string    `json:"user_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Channel   Channel   `json:"channel"`
	Limit     int       `json:"limit,omitempty"`
}

// StatementResponse represents statement response
type StatementResponse struct {
	AccountID    string        `json:"account_id"`
	StartDate    time.Time     `json:"start_date"`
	EndDate      time.Time     `json:"end_date"`
	Transactions []Transaction `json:"transactions"`
	Count        int           `json:"count"`
	GeneratedAt  time.Time     `json:"generated_at"`
}

// TransferRequest represents a fund transfer request
type TransferRequest struct {
	UserID      string          `json:"user_id"`
	FromAccount string          `json:"from_account"`
	ToAccount   string          `json:"to_account"`
	IFSC        string          `json:"ifsc,omitempty"`
	Amount      float64         `json:"amount"`
	Type        TransactionType `json:"type"`
	Remarks     string          `json:"remarks,omitempty"`
	Channel     Channel         `json:"channel"`
}

// TransferResponse represents transfer response
type TransferResponse struct {
	TransactionID   string    `json:"transaction_id"`
	Status          string    `json:"status"`
	Amount          float64   `json:"amount"`
	FromAccount     string    `json:"from_account"`
	ToAccount       string    `json:"to_account"`
	ReferenceNumber string    `json:"reference_number"`
	ProcessedAt     time.Time `json:"processed_at"`
	Message         string    `json:"message"`
}

// BalanceRequest represents balance inquiry request
type BalanceRequest struct {
	UserID    string `json:"user_id"`
	AccountID string `json:"account_id"`
	Channel   Channel `json:"channel"`
}

// BalanceResponse represents balance response
type BalanceResponse struct {
	AccountID   string    `json:"account_id"`
	AccountNumber string  `json:"account_number"`
	Balance     float64   `json:"balance"`
	Currency    string    `json:"currency"`
	AvailableBalance float64 `json:"available_balance"`
	LastUpdated time.Time `json:"last_updated"`
}

// DWHQueryRequest represents DWH query request
type DWHQueryRequest struct {
	QueryType string                 `json:"query_type"` // TRANSACTION_HISTORY, USER_PROFILE, ANALYTICS
	UserID    string                 `json:"user_id,omitempty"`
	AccountID string                 `json:"account_id,omitempty"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	StartDate *time.Time             `json:"start_date,omitempty"`
	EndDate   *time.Time             `json:"end_date,omitempty"`
	Limit     int                    `json:"limit,omitempty"`
}

// DWHQueryResponse represents DWH query response
type DWHQueryResponse struct {
	QueryType  string                   `json:"query_type"`
	Data       []map[string]interface{} `json:"data"`
	Count      int                      `json:"count"`
	ExecutedAt time.Time                `json:"executed_at"`
}

