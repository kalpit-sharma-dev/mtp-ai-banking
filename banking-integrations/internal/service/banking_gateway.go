package service

import (
	"context"
	"fmt"

	"github.com/aibanking/banking-integrations/internal/model"
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
	switch req.Channel {
	case model.ChannelMB:
		return bg.mbService.GetBalance(ctx, req)
	case model.ChannelNB:
		return bg.nbService.GetBalance(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported channel: %s", req.Channel)
	}
}

// TransferFunds processes transfer based on channel
func (bg *BankingGateway) TransferFunds(ctx context.Context, req *model.TransferRequest) (*model.TransferResponse, error) {
	switch req.Channel {
	case model.ChannelMB:
		return bg.mbService.TransferFunds(ctx, req)
	case model.ChannelNB:
		return bg.nbService.TransferFunds(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported channel: %s", req.Channel)
	}
}

// GetStatement retrieves statement based on channel
func (bg *BankingGateway) GetStatement(ctx context.Context, req *model.StatementRequest) (*model.StatementResponse, error) {
	switch req.Channel {
	case model.ChannelMB:
		return bg.mbService.GetStatement(ctx, req)
	case model.ChannelNB:
		return bg.nbService.GetStatement(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported channel: %s", req.Channel)
	}
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

