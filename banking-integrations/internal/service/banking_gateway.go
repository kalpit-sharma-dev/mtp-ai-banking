package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aibanking/banking-integrations/internal/model"
	"github.com/rs/zerolog/log"
)

// BankingGateway provides unified interface for all banking channels
type BankingGateway struct {
	mbService  *MBService
	nbService  *NBService
	dwhService *DWHService
}

// NewBankingGateway creates a new banking gateway
func NewBankingGateway(mbService *MBService, nbService *NBService, dwhService *DWHService) *BankingGateway {
	return &BankingGateway{
		mbService:  mbService,
		nbService:  nbService,
		dwhService: dwhService,
	}
}

// GetBalance retrieves balance based on channel
func (bg *BankingGateway) GetBalance(ctx context.Context, req *model.BalanceRequest) (*model.BalanceResponse, error) {
	// Get balance from DWH (which maintains actual balance)
	balance := bg.dwhService.GetBalance(req.UserID)
	availableBalance := balance - 5000.0 // Reserve for pending transactions
	
	// Get response from channel service (for other fields)
	var response *model.BalanceResponse
	var err error
	switch req.Channel {
	case model.ChannelMB:
		response, err = bg.mbService.GetBalance(ctx, req)
	case model.ChannelNB:
		response, err = bg.nbService.GetBalance(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported channel: %s", req.Channel)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Update with actual balance from DWH
	response.Balance = balance
	response.AvailableBalance = availableBalance
	response.LastUpdated = time.Now()
	
	return response, nil
}

// TransferFunds processes transfer based on channel
func (bg *BankingGateway) TransferFunds(ctx context.Context, req *model.TransferRequest) (*model.TransferResponse, error) {
	// Get current balance
	currentBalance := bg.dwhService.GetBalance(req.UserID)
	
	// Check if sufficient balance
	if currentBalance < req.Amount {
		return nil, fmt.Errorf("insufficient balance. Available: %.2f, Required: %.2f", currentBalance, req.Amount)
	}
	
	// Process transfer via channel service
	var response *model.TransferResponse
	var err error
	switch req.Channel {
	case model.ChannelMB:
		response, err = bg.mbService.TransferFunds(ctx, req)
	case model.ChannelNB:
		response, err = bg.nbService.TransferFunds(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported channel: %s", req.Channel)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Update balance after successful transfer
	newBalance := currentBalance - req.Amount
	bg.dwhService.UpdateBalance(req.UserID, newBalance)
	
	// Store transaction in DWH
	transaction := model.Transaction{
		TransactionID: response.TransactionID,
		AccountID:     req.FromAccount,
		UserID:        req.UserID,
		Type:          req.Type,
		Amount:        req.Amount,
		Currency:      "INR",
		FromAccount:   req.FromAccount,
		ToAccount:     req.ToAccount,
		IFSC:          req.IFSC,
		Status:        model.TransactionStatusCompleted,
		Channel:       req.Channel,
		ReferenceNumber: response.ReferenceNumber,
		CreatedAt:     response.ProcessedAt,
		CompletedAt:   &response.ProcessedAt,
	}
	bg.dwhService.StoreTransaction(req.UserID, transaction)
	
	log.Info().
		Str("user_id", req.UserID).
		Str("transaction_id", response.TransactionID).
		Float64("amount", req.Amount).
		Float64("old_balance", currentBalance).
		Float64("new_balance", newBalance).
		Msg("Balance updated after transfer")
	
	return response, nil
}

// GetStatement retrieves statement based on channel
func (bg *BankingGateway) GetStatement(ctx context.Context, req *model.StatementRequest) (*model.StatementResponse, error) {
	log.Info().
		Str("user_id", req.UserID).
		Time("start_date", req.StartDate).
		Time("end_date", req.EndDate).
		Int("limit", req.Limit).
		Msg("BankingGateway: Getting statement")
	
	// Get transactions from DWH (last 90 days)
	allTransactions, err := bg.dwhService.GetTransactionHistory(ctx, req.UserID, 90)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", req.UserID).
			Msg("Failed to get transaction history from DWH")
		return nil, err
	}
	
	log.Info().
		Str("user_id", req.UserID).
		Int("total_from_dwh", len(allTransactions)).
		Msg("BankingGateway: Retrieved transactions from DWH")
	
	// Filter transactions by date range
	filteredTransactions := []model.Transaction{}
	for _, txn := range allTransactions {
		if (txn.CreatedAt.After(req.StartDate) || txn.CreatedAt.Equal(req.StartDate)) &&
			(txn.CreatedAt.Before(req.EndDate) || txn.CreatedAt.Equal(req.EndDate)) {
			filteredTransactions = append(filteredTransactions, txn)
		}
	}
	
	log.Info().
		Str("user_id", req.UserID).
		Int("after_date_filter", len(filteredTransactions)).
		Msg("BankingGateway: Filtered transactions by date range")
	
	// Sort by date (newest first)
	sort.Slice(filteredTransactions, func(i, j int) bool {
		return filteredTransactions[i].CreatedAt.After(filteredTransactions[j].CreatedAt)
	})
	
	// Apply limit if specified
	if req.Limit > 0 && len(filteredTransactions) > req.Limit {
		filteredTransactions = filteredTransactions[:req.Limit]
	}
	
	// Create response directly with DWH transactions (don't use channel service mock data)
	response := &model.StatementResponse{
		AccountID:    req.AccountID,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Transactions: filteredTransactions,
		Count:        len(filteredTransactions),
		GeneratedAt:  time.Now(),
	}
	
	log.Info().
		Str("user_id", req.UserID).
		Int("final_count", len(filteredTransactions)).
		Msg("BankingGateway: Returning statement with actual transactions")
	
	return response, nil
}

// AddBeneficiary adds beneficiary based on channel
func (bg *BankingGateway) AddBeneficiary(ctx context.Context, channel model.Channel, userID, accountNumber, ifsc, name string) (*model.Beneficiary, error) {
	var beneficiary *model.Beneficiary
	var err error
	
	switch channel {
	case model.ChannelMB:
		beneficiary, err = bg.mbService.AddBeneficiary(ctx, userID, accountNumber, ifsc, name)
	case model.ChannelNB:
		beneficiary, err = bg.nbService.AddBeneficiary(ctx, userID, accountNumber, ifsc, name)
	default:
		return nil, fmt.Errorf("unsupported channel: %s", channel)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Store beneficiary in DWH for retrieval
	beneficiaryMap := map[string]interface{}{
		"beneficiary_id": beneficiary.BeneficiaryID,
		"user_id":        beneficiary.UserID,
		"account_number": beneficiary.AccountNumber,
		"ifsc":           beneficiary.IFSC,
		"name":           beneficiary.Name,
		"account_type":   beneficiary.AccountType,
		"status":         beneficiary.Status,
		"added_at":       beneficiary.AddedAt,
	}
	bg.dwhService.StoreBeneficiary(userID, beneficiaryMap)
	
	return beneficiary, nil
}

// QueryDWH queries data warehouse
func (bg *BankingGateway) QueryDWH(ctx context.Context, req *model.DWHQueryRequest) (*model.DWHQueryResponse, error) {
	return bg.dwhService.Query(ctx, req)
}

// GetTransactionHistory retrieves transaction history from DWH
func (bg *BankingGateway) GetTransactionHistory(ctx context.Context, userID string, days int) ([]model.Transaction, error) {
	return bg.dwhService.GetTransactionHistory(ctx, userID, days)
}

