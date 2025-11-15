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
	switch channel {
	case model.ChannelMB:
		return bg.mbService.AddBeneficiary(ctx, userID, accountNumber, ifsc, name)
	case model.ChannelNB:
		return bg.nbService.AddBeneficiary(ctx, userID, accountNumber, ifsc, name)
	default:
		return nil, fmt.Errorf("unsupported channel: %s", channel)
	}
}

// QueryDWH queries data warehouse
func (bg *BankingGateway) QueryDWH(ctx context.Context, req *model.DWHQueryRequest) (*model.DWHQueryResponse, error) {
	return bg.dwhService.Query(ctx, req)
}

// GetTransactionHistory retrieves transaction history from DWH
func (bg *BankingGateway) GetTransactionHistory(ctx context.Context, userID string, days int) ([]model.Transaction, error) {
	return bg.dwhService.GetTransactionHistory(ctx, userID, days)
}

