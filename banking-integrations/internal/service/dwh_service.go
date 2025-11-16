package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aibanking/banking-integrations/internal/config"
	"github.com/aibanking/banking-integrations/internal/model"
	"github.com/rs/zerolog/log"
)

// DWHService handles Data Warehouse operations
type DWHService struct {
	config        *config.DWHConfig
	beneficiaries map[string][]map[string]interface{} // In-memory storage: userID -> beneficiaries
	transactions  map[string][]model.Transaction      // In-memory storage: userID -> transactions
	balances      map[string]float64                   // In-memory storage: userID -> balance
	mu            sync.RWMutex                         // Mutex for thread-safe access
	// In production, this would have database connection
}

// NewDWHService creates a new DWH service
func NewDWHService(cfg *config.DWHConfig) *DWHService {
	// Initialize with default balance
	balances := make(map[string]float64)
	balances["U10001"] = 150000.0 // Default balance for demo user
	
	return &DWHService{
		config:        cfg,
		beneficiaries: make(map[string][]map[string]interface{}),
		transactions:  make(map[string][]model.Transaction),
		balances:      balances,
	}
}

// Query executes a query against the data warehouse
func (dwh *DWHService) Query(ctx context.Context, req *model.DWHQueryRequest) (*model.DWHQueryResponse, error) {
	log.Info().
		Str("query_type", req.QueryType).
		Str("user_id", req.UserID).
		Msg("DWH: Executing query")

	var data []map[string]interface{}

	switch req.QueryType {
	case "TRANSACTION_HISTORY":
		data = dwh.getTransactionHistory(ctx, req)
	case "USER_PROFILE":
		data = dwh.getUserProfile(ctx, req)
	case "ANALYTICS":
		data = dwh.getAnalytics(ctx, req)
	case "BENEFICIARIES":
		data = dwh.getBeneficiaries(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported query type: %s", req.QueryType)
	}

	return &model.DWHQueryResponse{
		QueryType:  req.QueryType,
		Data:       data,
		Count:      len(data),
		ExecutedAt: time.Now(),
	}, nil
}

// getTransactionHistory retrieves transaction history from DWH
func (dwh *DWHService) getTransactionHistory(ctx context.Context, req *model.DWHQueryRequest) []map[string]interface{} {
	dwh.mu.RLock()
	defer dwh.mu.RUnlock()
	
	// Get stored transactions for this user
	if transactions, exists := dwh.transactions[req.UserID]; exists {
		history := []map[string]interface{}{}
		for _, txn := range transactions {
			history = append(history, map[string]interface{}{
				"transaction_id": txn.TransactionID,
				"user_id":        txn.UserID,
				"amount":         txn.Amount,
				"type":           string(txn.Type),
				"status":         string(txn.Status),
				"from_account":   txn.FromAccount,
				"to_account":     txn.ToAccount,
				"created_at":     txn.CreatedAt,
			})
		}
		return history
	}
	
	// Return empty if no transactions
	return []map[string]interface{}{}
}

// getUserProfile retrieves user profile from DWH
func (dwh *DWHService) getUserProfile(ctx context.Context, req *model.DWHQueryRequest) []map[string]interface{} {
	// Mock data - in production would query DWH database
	profile := []map[string]interface{}{
		{
			"user_id":          req.UserID,
			"account_age_days": 365,
			"total_balance":    150000.0,
			"monthly_income":   50000.0,
			"transaction_count_30d": 25,
			"avg_transaction_amount": 10000.0,
			"credit_score":    750,
			"kyc_status":      "VERIFIED",
			"account_type":    "SAVINGS",
		},
	}

	return profile
}

// getAnalytics retrieves analytics data from DWH
func (dwh *DWHService) getAnalytics(ctx context.Context, req *model.DWHQueryRequest) []map[string]interface{} {
	// Mock analytics - in production would query DWH database
	analytics := []map[string]interface{}{
		{
			"metric":          "total_transactions",
			"value":           150,
			"period":          "30d",
			"calculated_at":   time.Now(),
		},
		{
			"metric":          "total_amount",
			"value":           1500000.0,
			"period":          "30d",
			"calculated_at":   time.Now(),
		},
		{
			"metric":          "avg_transaction_amount",
			"value":           10000.0,
			"period":          "30d",
			"calculated_at":   time.Now(),
		},
		{
			"metric":          "fraud_rate",
			"value":           0.02,
			"period":          "30d",
			"calculated_at":   time.Now(),
		},
	}

	return analytics
}

// GetTransactionHistory retrieves transaction history for a user
func (dwh *DWHService) GetTransactionHistory(ctx context.Context, userID string, days int) ([]model.Transaction, error) {
	dwh.mu.RLock()
	defer dwh.mu.RUnlock()
	
	log.Info().
		Str("user_id", userID).
		Int("days", days).
		Msg("DWH: Getting transaction history")

	// Return stored transactions for this user
	if transactions, exists := dwh.transactions[userID]; exists {
		log.Info().
			Str("user_id", userID).
			Int("total_stored", len(transactions)).
			Msg("DWH: Found stored transactions")
		
		// Filter by date if needed
		cutoffDate := time.Now().AddDate(0, 0, -days)
		filtered := []model.Transaction{}
		for _, txn := range transactions {
			if txn.CreatedAt.After(cutoffDate) || txn.CreatedAt.Equal(cutoffDate) {
				filtered = append(filtered, txn)
			}
		}
		
		log.Info().
			Str("user_id", userID).
			Int("filtered_count", len(filtered)).
			Time("cutoff_date", cutoffDate).
			Msg("DWH: Returning filtered transactions")
		
		return filtered, nil
	}

	// Return empty if no transactions
	log.Info().
		Str("user_id", userID).
		Msg("DWH: No transactions found for user")
	return []model.Transaction{}, nil
}

// getBeneficiaries retrieves beneficiaries for a user
func (dwh *DWHService) getBeneficiaries(ctx context.Context, req *model.DWHQueryRequest) []map[string]interface{} {
	// Return stored beneficiaries for this user
	if beneficiaries, exists := dwh.beneficiaries[req.UserID]; exists {
		return beneficiaries
	}
	// Return empty array if no beneficiaries found
	return []map[string]interface{}{}
}

// StoreBeneficiary stores a beneficiary in memory
func (dwh *DWHService) StoreBeneficiary(userID string, beneficiary map[string]interface{}) {
	if dwh.beneficiaries == nil {
		dwh.beneficiaries = make(map[string][]map[string]interface{})
	}
	dwh.beneficiaries[userID] = append(dwh.beneficiaries[userID], beneficiary)
}

// StoreTransaction stores a transaction in memory
func (dwh *DWHService) StoreTransaction(userID string, transaction model.Transaction) {
	dwh.mu.Lock()
	defer dwh.mu.Unlock()
	
	if dwh.transactions == nil {
		dwh.transactions = make(map[string][]model.Transaction)
	}
	dwh.transactions[userID] = append(dwh.transactions[userID], transaction)
	
	log.Info().
		Str("user_id", userID).
		Str("transaction_id", transaction.TransactionID).
		Float64("amount", transaction.Amount).
		Str("type", string(transaction.Type)).
		Int("total_transactions", len(dwh.transactions[userID])).
		Msg("DWH: Transaction stored")
}

// GetBalance retrieves balance for a user
func (dwh *DWHService) GetBalance(userID string) float64 {
	dwh.mu.RLock()
	defer dwh.mu.RUnlock()
	
	if balance, exists := dwh.balances[userID]; exists {
		log.Debug().
			Str("user_id", userID).
			Float64("balance", balance).
			Msg("DWH: Retrieved balance")
		return balance
	}
	// Default balance for new users
	log.Debug().
		Str("user_id", userID).
		Float64("default_balance", 150000.0).
		Msg("DWH: Using default balance")
	return 150000.0
}

// UpdateBalance updates balance for a user
func (dwh *DWHService) UpdateBalance(userID string, newBalance float64) {
	dwh.mu.Lock()
	defer dwh.mu.Unlock()
	
	if dwh.balances == nil {
		dwh.balances = make(map[string]float64)
	}
	
	oldBalance := dwh.balances[userID]
	dwh.balances[userID] = newBalance
	
	log.Info().
		Str("user_id", userID).
		Float64("old_balance", oldBalance).
		Float64("new_balance", newBalance).
		Float64("difference", newBalance-oldBalance).
		Msg("DWH: Balance updated")
}

