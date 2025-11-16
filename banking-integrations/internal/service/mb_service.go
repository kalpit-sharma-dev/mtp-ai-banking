package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aibanking/banking-integrations/internal/model"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// MBService handles Mobile Banking operations
type MBService struct {
	// In production, this would have database connection
}

// NewMBService creates a new MB service
func NewMBService() *MBService {
	return &MBService{}
}

// GetBalance retrieves account balance for mobile banking
// Note: This method should receive dwhService to get actual balance, but for now we'll use a shared approach
func (mb *MBService) GetBalance(ctx context.Context, req *model.BalanceRequest) (*model.BalanceResponse, error) {
	log.Info().
		Str("user_id", req.UserID).
		Str("account_id", req.AccountID).
		Str("channel", string(req.Channel)).
		Msg("MB: Getting balance")

	// Mock implementation - in production would query database
	// For now, use default balance (will be updated by gateway)
	balance := 150000.0
	availableBalance := balance - 5000.0 // Reserve for pending transactions

	return &model.BalanceResponse{
		AccountID:        req.AccountID,
		AccountNumber:    "XXXX1234",
		Balance:          balance,
		Currency:         "INR",
		AvailableBalance: availableBalance,
		LastUpdated:      time.Now(),
	}, nil
}

// TransferFunds processes fund transfer via mobile banking
func (mb *MBService) TransferFunds(ctx context.Context, req *model.TransferRequest) (*model.TransferResponse, error) {
	log.Info().
		Str("user_id", req.UserID).
		Str("from_account", req.FromAccount).
		Str("to_account", req.ToAccount).
		Float64("amount", req.Amount).
		Str("type", string(req.Type)).
		Msg("MB: Processing fund transfer")

	// Generate transaction ID
	txnID := fmt.Sprintf("MB_%s", uuid.New().String()[:8])
	refNumber := fmt.Sprintf("REF%s", uuid.New().String()[:12])

	// Mock processing - in production would call core banking system
	status := "COMPLETED"
	message := "Transfer processed successfully"

	return &model.TransferResponse{
		TransactionID:   txnID,
		Status:          status,
		Amount:          req.Amount,
		FromAccount:     req.FromAccount,
		ToAccount:       req.ToAccount,
		ReferenceNumber: refNumber,
		ProcessedAt:     time.Now(),
		Message:         message,
	}, nil
}

// GetStatement retrieves account statement for mobile banking
func (mb *MBService) GetStatement(ctx context.Context, req *model.StatementRequest) (*model.StatementResponse, error) {
	log.Info().
		Str("account_id", req.AccountID).
		Str("user_id", req.UserID).
		Time("start_date", req.StartDate).
		Time("end_date", req.EndDate).
		Msg("MB: Getting statement")

	// Mock transactions - in production would query database
	transactions := []model.Transaction{
		{
			TransactionID: "TXN_001",
			AccountID:     req.AccountID,
			UserID:        req.UserID,
			Type:          model.TransactionTypeDEBIT,
			Amount:        25000.0,
			Currency:      "INR",
			Status:        model.TransactionStatusCompleted,
			Channel:       model.ChannelMB,
			CreatedAt:     time.Now().AddDate(0, 0, -5),
		},
		{
			TransactionID: "TXN_002",
			AccountID:     req.AccountID,
			UserID:        req.UserID,
			Type:          model.TransactionTypeCREDIT,
			Amount:        50000.0,
			Currency:      "INR",
			Status:        model.TransactionStatusCompleted,
			Channel:       model.ChannelMB,
			CreatedAt:     time.Now().AddDate(0, 0, -10),
		},
	}

	return &model.StatementResponse{
		AccountID:    req.AccountID,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Transactions: transactions,
		Count:        len(transactions),
		GeneratedAt:  time.Now(),
	}, nil
}

// AddBeneficiary adds a beneficiary for mobile banking
func (mb *MBService) AddBeneficiary(ctx context.Context, userID, accountNumber, ifsc, name string) (*model.Beneficiary, error) {
	log.Info().
		Str("user_id", userID).
		Str("account_number", accountNumber).
		Str("ifsc", ifsc).
		Msg("MB: Adding beneficiary")

	beneficiaryID := fmt.Sprintf("BEN_%s", uuid.New().String()[:8])

	return &model.Beneficiary{
		BeneficiaryID: beneficiaryID,
		UserID:        userID,
		AccountNumber: accountNumber,
		IFSC:          ifsc,
		Name:          name,
		AccountType:   "SAVINGS",
		Status:        "ACTIVE",
		AddedAt:       time.Now(),
	}, nil
}

