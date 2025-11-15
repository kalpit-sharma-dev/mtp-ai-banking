package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aibanking/agent-mesh/internal/model"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// BankingAgent handles banking operations
type BankingAgent struct {
	*AgentBase
}

// NewBankingAgent creates a new banking agent
func NewBankingAgent(base *AgentBase) *BankingAgent {
	return &BankingAgent{
		AgentBase: base,
	}
}

// Process processes a banking request
func (ba *BankingAgent) Process(ctx context.Context, req *model.AgentRequest) (*model.AgentResponse, error) {
	log.Info().
		Str("task", req.Task).
		Str("request_id", req.RequestID).
		Msg("Banking agent processing request")

	inputCtx := req.InputContext
	task := req.Task

	switch task {
	case "TRANSFER_NEFT", "TRANSFER_RTGS", "TRANSFER_IMPS", "TRANSFER_UPI":
		return ba.processTransfer(ctx, req, inputCtx)
	case "CHECK_BALANCE":
		return ba.checkBalance(ctx, req, inputCtx)
	case "GET_STATEMENT":
		return ba.getStatement(ctx, req, inputCtx)
	case "ADD_BENEFICIARY":
		return ba.addBeneficiary(ctx, req, inputCtx)
	default:
		return &model.AgentResponse{
			AgentID:     ba.agentType,
			AgentType:   "BANKING",
			Status:      "REJECTED",
			Result:      map[string]interface{}{"error": "Unsupported operation"},
			RiskScore:   0.0,
			Explanation: fmt.Sprintf("Banking agent does not support operation: %s", task),
			Confidence:  1.0,
			Timestamp:   time.Now(),
			RequestID:   req.RequestID,
		}, nil
	}
}

// processTransfer processes a fund transfer
func (ba *BankingAgent) processTransfer(ctx context.Context, req *model.AgentRequest, inputCtx map[string]interface{}) (*model.AgentResponse, error) {
	data, ok := inputCtx["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data in input context")
	}

	amount, ok := data["amount"].(float64)
	if !ok {
		return nil, fmt.Errorf("amount not found or invalid")
	}

	toAccount, ok := data["to_account"].(string)
	if !ok {
		return nil, fmt.Errorf("to_account not found")
	}

	// Generate transaction ID
	txnID := fmt.Sprintf("TXN_%s", uuid.New().String()[:8])

	// Simulate transfer processing
	log.Info().
		Float64("amount", amount).
		Str("to_account", toAccount).
		Str("txn_id", txnID).
		Msg("Processing fund transfer")

	// In production, this would call actual banking core system
	result := map[string]interface{}{
		"status":          "APPROVED",
		"transaction_id":  txnID,
		"amount":          amount,
		"to_account":      toAccount,
		"message":         "Transfer processed successfully",
		"processed_at":    time.Now(),
	}

	return &model.AgentResponse{
		AgentID:     ba.agentType,
		AgentType:   "BANKING",
		Status:      "APPROVED",
		Result:      result,
		RiskScore:   0.1,
		Explanation: "Fund transfer processed successfully within banking limits",
		Confidence:  0.95,
		Timestamp:   time.Now(),
		RequestID:   req.RequestID,
	}, nil
}

// checkBalance checks account balance
func (ba *BankingAgent) checkBalance(ctx context.Context, req *model.AgentRequest, inputCtx map[string]interface{}) (*model.AgentResponse, error) {
	userID, _ := inputCtx["user_id"].(string)

	// Mock balance - in production would query database
	balance := 150000.0

	log.Info().
		Str("user_id", userID).
		Float64("balance", balance).
		Msg("Checking account balance")

	result := map[string]interface{}{
		"balance":   balance,
		"currency":  "INR",
		"account_id": userID,
		"checked_at": time.Now(),
	}

	return &model.AgentResponse{
		AgentID:     ba.agentType,
		AgentType:   "BANKING",
		Status:      "APPROVED",
		Result:      result,
		RiskScore:   0.0,
		Explanation: "Balance retrieved successfully",
		Confidence:  1.0,
		Timestamp:   time.Now(),
		RequestID:   req.RequestID,
	}, nil
}

// getStatement retrieves account statement
func (ba *BankingAgent) getStatement(ctx context.Context, req *model.AgentRequest, inputCtx map[string]interface{}) (*model.AgentResponse, error) {
	userID, _ := inputCtx["user_id"].(string)

	// Mock statement - in production would query database
	transactions := []map[string]interface{}{
		{
			"transaction_id": "TXN_001",
			"type":           "DEBIT",
			"amount":         25000.0,
			"description":    "NEFT Transfer",
			"date":           time.Now().AddDate(0, 0, -5),
		},
		{
			"transaction_id": "TXN_002",
			"type":           "CREDIT",
			"amount":         50000.0,
			"description":    "Salary Credit",
			"date":           time.Now().AddDate(0, 0, -10),
		},
	}

	result := map[string]interface{}{
		"account_id":    userID,
		"transactions":  transactions,
		"count":         len(transactions),
		"generated_at":  time.Now(),
	}

	return &model.AgentResponse{
		AgentID:     ba.agentType,
		AgentType:   "BANKING",
		Status:      "APPROVED",
		Result:      result,
		RiskScore:   0.0,
		Explanation: "Statement generated successfully",
		Confidence:  1.0,
		Timestamp:   time.Now(),
		RequestID:   req.RequestID,
	}, nil
}

// addBeneficiary adds a new beneficiary
func (ba *BankingAgent) addBeneficiary(ctx context.Context, req *model.AgentRequest, inputCtx map[string]interface{}) (*model.AgentResponse, error) {
	data, ok := inputCtx["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data in input context")
	}

	account, _ := data["account"].(string)
	name, _ := data["name"].(string)
	ifsc, _ := data["ifsc"].(string)

	log.Info().
		Str("account", account).
		Str("name", name).
		Msg("Adding beneficiary")

	result := map[string]interface{}{
		"status":         "APPROVED",
		"beneficiary_id": fmt.Sprintf("BEN_%s", uuid.New().String()[:8]),
		"account":        account,
		"name":           name,
		"ifsc":           ifsc,
		"added_at":       time.Now(),
	}

	return &model.AgentResponse{
		AgentID:     ba.agentType,
		AgentType:   "BANKING",
		Status:      "APPROVED",
		Result:      result,
		RiskScore:   0.1,
		Explanation: "Beneficiary added successfully",
		Confidence:  0.9,
		Timestamp:   time.Now(),
		RequestID:   req.RequestID,
	}, nil
}

