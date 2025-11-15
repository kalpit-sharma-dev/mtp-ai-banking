package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aibanking/banking-integrations/internal/model"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// NBService handles Net Banking operations
type NBService struct {
	// In production, this would have database connection
}

// NewNBService creates a new NB service
func NewNBService() *NBService {
	return &NBService{}
}

// GetBalance retrieves account balance for net banking
func (nb *NBService) GetBalance(ctx context.Context, req *model.BalanceRequest) (*model.BalanceResponse, error) {
	log.Info().
		Str("user_id", req.UserID).
		Str("account_id", req.AccountID).
		Str("channel", string(req.Channel)).
		Msg("NB: Getting balance")

	// Mock implementation - in production would query database
	balance := 150000.0
	availableBalance := balance - 5000.0

	return &model.BalanceResponse{
		AccountID:        req.AccountID,
		AccountNumber:    "XXXX1234",
		Balance:          balance,
		Currency:         "INR",
		AvailableBalance: availableBalance,
		LastUpdated:      time.Now(),
	}, nil
}

// TransferFunds processes fund transfer via net banking
func (nb *NBService) TransferFunds(ctx context.Context, req *model.TransferRequest) (*model.TransferResponse, error) {
	log.Info().
		Str("user_id", req.UserID).
		Str("from_account", req.FromAccount).
		Str("to_account", req.ToAccount).
		Float64("amount", req.Amount).
		Str("type", string(req.Type)).
		Msg("NB: Processing fund transfer")

	// Generate transaction ID
	txnID := fmt.Sprintf("NB_%s", uuid.New().String()[:8])
	refNumber := fmt.Sprintf("REF%s", uuid.New().String()[:12])

	// Mock processing
	status := "COMPLETED"
	message := "Transfer processed successfully via Net Banking"

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

// GetStatement retrieves account statement for net banking
func (nb *NBService) GetStatement(ctx context.Context, req *model.StatementRequest) (*model.StatementResponse, error) {
	log.Info().
		Str("account_id", req.AccountID).
		Str("user_id", req.UserID).
		Time("start_date", req.StartDate).
		Time("end_date", req.EndDate).
		Msg("NB: Getting statement")

	// Mock transactions
	transactions := []model.Transaction{
		{
			TransactionID: "TXN_003",
			AccountID:     req.AccountID,
			UserID:        req.UserID,
			Type:          model.TransactionTypeNEFT,
			Amount:        50000.0,
			Currency:      "INR",
			ToAccount:     "YYYY5678",
			Status:        model.TransactionStatusCompleted,
			Channel:       model.ChannelNB,
			CreatedAt:     time.Now().AddDate(0, 0, -3),
		},
		{
			TransactionID: "TXN_004",
			AccountID:     req.AccountID,
			UserID:        req.UserID,
			Type:          model.TransactionTypeRTGS,
			Amount:        100000.0,
			Currency:      "INR",
			ToAccount:     "ZZZZ9012",
			Status:        model.TransactionStatusCompleted,
			Channel:       model.ChannelNB,
			CreatedAt:     time.Now().AddDate(0, 0, -7),
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

// AddBeneficiary adds a beneficiary for net banking
func (nb *NBService) AddBeneficiary(ctx context.Context, userID, accountNumber, ifsc, name string) (*model.Beneficiary, error) {
	log.Info().
		Str("user_id", userID).
		Str("account_number", accountNumber).
		Str("ifsc", ifsc).
		Msg("NB: Adding beneficiary")

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

