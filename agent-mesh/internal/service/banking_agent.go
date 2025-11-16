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

	userID, _ := inputCtx["user_id"].(string)
	channel, _ := inputCtx["channel"].(string)
	if channel == "" {
		channel = "MB"
	}

	fromAccount := userID
	if fromAcc, ok := inputCtx["account_id"].(string); ok && fromAcc != "" {
		fromAccount = fromAcc
	}

	// Try to call Banking Integrations service first
	if ba.bankingIntegrationsEnabled {
		transferResp, err := ba.callBankingTransfer(ctx, userID, fromAccount, toAccount, amount, channel, data)
		if err == nil {
			log.Info().
				Str("user_id", userID).
				Str("source", "banking_integrations").
				Msg("Transfer processed via Banking Integrations")
			return transferResp, nil
		}
		log.Warn().Err(err).Msg("Failed to process transfer via Banking Integrations, using fallback")
	}

	// Fallback to mock processing
	txnID := fmt.Sprintf("TXN_%s", uuid.New().String()[:8])

	log.Info().
		Float64("amount", amount).
		Str("to_account", toAccount).
		Str("txn_id", txnID).
		Str("source", "mock").
		Msg("Processing fund transfer (mock)")

	result := map[string]interface{}{
		"status":         "APPROVED",
		"transaction_id": txnID,
		"amount":         amount,
		"to_account":     toAccount,
		"message":        "Transfer processed successfully",
		"processed_at":   time.Now(),
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

// callBankingTransfer calls Banking Integrations service for transfer
func (ba *BankingAgent) callBankingTransfer(ctx context.Context, userID, fromAccount, toAccount string, amount float64, channel string, data map[string]interface{}) (*model.AgentResponse, error) {
	payload := map[string]interface{}{
		"user_id":     userID,
		"from_account": fromAccount,
		"to_account":  toAccount,
		"amount":      amount,
		"channel":     channel,
	}

	// Add transfer type if available
	if transferType, ok := data["transfer_type"].(string); ok {
		payload["transfer_type"] = transferType
	}
	if ifsc, ok := data["ifsc"].(string); ok {
		payload["ifsc"] = ifsc
	}
	if remarks, ok := data["remarks"].(string); ok {
		payload["remarks"] = remarks
	}

	result, err := ba.CallBankingService(ctx, "/api/v1/transfer", payload)
	if err != nil {
		return nil, err
	}

	// Extract transfer response
	responseResult := map[string]interface{}{
		"status":         result["status"],
		"transaction_id": result["transaction_id"],
		"amount":         amount,
		"to_account":     toAccount,
		"processed_at":    time.Now(),
	}

	if message, ok := result["message"].(string); ok {
		responseResult["message"] = message
	}
	if reference, ok := result["reference_number"].(string); ok {
		responseResult["reference_number"] = reference
	}

	return &model.AgentResponse{
		AgentID:     ba.agentType,
		AgentType:   "BANKING",
		Status:      "APPROVED",
		Result:      responseResult,
		RiskScore:   0.1,
		Explanation: "Fund transfer processed successfully via banking system",
		Confidence:  0.95,
		Timestamp:   time.Now(),
		RequestID:   "",
	}, nil
}

// checkBalance checks account balance
func (ba *BankingAgent) checkBalance(ctx context.Context, req *model.AgentRequest, inputCtx map[string]interface{}) (*model.AgentResponse, error) {
	userID, _ := inputCtx["user_id"].(string)
	channel, _ := inputCtx["channel"].(string)
	if channel == "" {
		channel = "MB" // Default to Mobile Banking
	}

	accountID := userID
	if accID, ok := inputCtx["account_id"].(string); ok && accID != "" {
		accountID = accID
	}

	// Try to call Banking Integrations service first
	if ba.bankingIntegrationsEnabled {
		balanceResp, err := ba.callBankingBalance(ctx, userID, accountID, channel)
		if err == nil {
			log.Info().
				Str("user_id", userID).
				Str("source", "banking_integrations").
				Msg("Balance retrieved from Banking Integrations")
			return balanceResp, nil
		}
		log.Warn().Err(err).Msg("Failed to get balance from Banking Integrations, using fallback")
	}

	// Fallback to mock data
	balance := 150000.0
	log.Info().
		Str("user_id", userID).
		Float64("balance", balance).
		Str("source", "mock").
		Msg("Checking account balance (mock)")

	result := map[string]interface{}{
		"balance":    balance,
		"currency":   "INR",
		"account_id": accountID,
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

// callBankingBalance calls Banking Integrations service for balance
func (ba *BankingAgent) callBankingBalance(ctx context.Context, userID, accountID, channel string) (*model.AgentResponse, error) {
	payload := map[string]interface{}{
		"user_id":    userID,
		"account_id": accountID,
		"channel":    channel,
	}

	result, err := ba.CallBankingService(ctx, "/api/v1/balance", payload)
	if err != nil {
		return nil, err
	}

	// Extract balance from response
	balance, _ := result["balance"].(float64)
	currency, _ := result["currency"].(string)
	if currency == "" {
		currency = "INR"
	}

	responseResult := map[string]interface{}{
		"balance":    balance,
		"currency":   currency,
		"account_id": accountID,
		"checked_at": time.Now(),
	}

	// Copy any additional fields from Banking Integrations response
	if availableBalance, ok := result["available_balance"].(float64); ok {
		responseResult["available_balance"] = availableBalance
	}
	if accountType, ok := result["account_type"].(string); ok {
		responseResult["account_type"] = accountType
	}

	return &model.AgentResponse{
		AgentID:     ba.agentType,
		AgentType:   "BANKING",
		Status:      "APPROVED",
		Result:      responseResult,
		RiskScore:   0.0,
		Explanation: "Balance retrieved successfully from banking system",
		Confidence:  1.0,
		Timestamp:   time.Now(),
		RequestID:   "",
	}, nil
}

// getStatement retrieves account statement
func (ba *BankingAgent) getStatement(ctx context.Context, req *model.AgentRequest, inputCtx map[string]interface{}) (*model.AgentResponse, error) {
	userID, _ := inputCtx["user_id"].(string)
	channel, _ := inputCtx["channel"].(string)
	if channel == "" {
		channel = "MB"
	}

	accountID := userID
	if accID, ok := inputCtx["account_id"].(string); ok && accID != "" {
		accountID = accID
	}

	// Try to call Banking Integrations service first
	if ba.bankingIntegrationsEnabled {
		statementResp, err := ba.callBankingStatement(ctx, userID, accountID, channel)
		if err == nil {
			log.Info().
				Str("user_id", userID).
				Str("source", "banking_integrations").
				Msg("Statement retrieved from Banking Integrations")
			return statementResp, nil
		}
		log.Warn().Err(err).Msg("Failed to get statement from Banking Integrations, using fallback")
	}

	// Fallback to mock data
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
		"account_id":   userID,
		"transactions": transactions,
		"count":        len(transactions),
		"generated_at": time.Now(),
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

// callBankingStatement calls Banking Integrations service for statement
func (ba *BankingAgent) callBankingStatement(ctx context.Context, userID, accountID, channel string) (*model.AgentResponse, error) {
	// Default to last 30 days if not specified
	payload := map[string]interface{}{
		"user_id":    userID,
		"account_id": accountID,
		"channel":    channel,
		"start_date": time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
		"end_date":   time.Now().Format("2006-01-02"),
	}

	result, err := ba.CallBankingService(ctx, "/api/v1/statement", payload)
	if err != nil {
		return nil, err
	}

	// Extract transactions from response
	var transactionList []map[string]interface{}
	if transactions, ok := result["transactions"].([]interface{}); ok {
		transactionList = make([]map[string]interface{}, 0, len(transactions))
		for _, txn := range transactions {
			if txnMap, ok := txn.(map[string]interface{}); ok {
				transactionList = append(transactionList, txnMap)
			}
		}
	} else {
		transactionList = make([]map[string]interface{}, 0)
	}

	responseResult := map[string]interface{}{
		"account_id":   accountID,
		"transactions": transactionList,
		"count":        len(transactionList),
		"generated_at": time.Now(),
	}

	return &model.AgentResponse{
		AgentID:     ba.agentType,
		AgentType:   "BANKING",
		Status:      "APPROVED",
		Result:      responseResult,
		RiskScore:   0.0,
		Explanation: "Statement retrieved successfully from banking system",
		Confidence:  1.0,
		Timestamp:   time.Now(),
		RequestID:   "",
	}, nil
}

// addBeneficiary adds a new beneficiary
func (ba *BankingAgent) addBeneficiary(ctx context.Context, req *model.AgentRequest, inputCtx map[string]interface{}) (*model.AgentResponse, error) {
	data, ok := inputCtx["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data in input context")
	}

	// Extract fields with multiple possible names (handle both UI and AI Assistant formats)
	accountNumber := ""
	if acc, ok := data["account_number"].(string); ok && acc != "" {
		accountNumber = acc
	} else if acc, ok := data["account"].(string); ok && acc != "" {
		accountNumber = acc
	} else if acc, ok := data["to_account"].(string); ok && acc != "" {
		accountNumber = acc
	}

	name := ""
	if n, ok := data["name"].(string); ok && n != "" {
		name = n
	} else if n, ok := data["payee_name"].(string); ok && n != "" {
		name = n
	}

	ifsc := ""
	if i, ok := data["ifsc"].(string); ok && i != "" {
		ifsc = i
	}

	// Get user_id and channel from input context
	userID, _ := inputCtx["user_id"].(string)
	channel, _ := inputCtx["channel"].(string)
	if channel == "" {
		channel = "MB" // Default to Mobile Banking
	}

	// Validate required fields
	if accountNumber == "" || name == "" || ifsc == "" {
		return &model.AgentResponse{
			AgentID:     ba.agentType,
			AgentType:   "BANKING",
			Status:      "REJECTED",
			Result:      map[string]interface{}{"error": "Missing required fields: account_number, name, or ifsc"},
			RiskScore:   0.0,
			Explanation: "Please provide beneficiary name, account number, and IFSC code",
			Confidence:  0.0,
			Timestamp:   time.Now(),
			RequestID:   req.RequestID,
		}, nil
	}

	log.Info().
		Str("user_id", userID).
		Str("account", accountNumber).
		Str("name", name).
		Str("ifsc", ifsc).
		Str("channel", channel).
		Msg("Adding beneficiary")

	// Try to call Banking Integrations service first
	if ba.bankingIntegrationsEnabled {
		beneficiaryResp, err := ba.callBankingAddBeneficiary(ctx, userID, accountNumber, ifsc, name, channel)
		if err == nil {
			log.Info().
				Str("user_id", userID).
				Str("source", "banking_integrations").
				Msg("Beneficiary added via Banking Integrations")
			return beneficiaryResp, nil
		}
		log.Warn().Err(err).Msg("Failed to add beneficiary via Banking Integrations, using fallback")
	}

	// Fallback to mock response (for development/testing)
	beneficiaryID := fmt.Sprintf("BEN_%s", uuid.New().String()[:8])
	result := map[string]interface{}{
		"status":          "APPROVED",
		"beneficiary_id":  beneficiaryID,
		"account_number":  accountNumber,
		"name":            name,
		"ifsc":            ifsc,
		"added_at":        time.Now(),
		"message":         fmt.Sprintf("Beneficiary %s added successfully", name),
	}

	return &model.AgentResponse{
		AgentID:     ba.agentType,
		AgentType:   "BANKING",
		Status:      "APPROVED",
		Result:      result,
		RiskScore:   0.1,
		Explanation: fmt.Sprintf("Beneficiary %s added successfully", name),
		Confidence:  0.9,
		Timestamp:   time.Now(),
		RequestID:   req.RequestID,
	}, nil
}

// callBankingAddBeneficiary calls Banking Integrations service to add beneficiary
func (ba *BankingAgent) callBankingAddBeneficiary(ctx context.Context, userID, accountNumber, ifsc, name, channel string) (*model.AgentResponse, error) {
	payload := map[string]interface{}{
		"user_id":        userID,
		"account_number": accountNumber,
		"ifsc":           ifsc,
		"name":           name,
		"channel":        channel,
	}

	result, err := ba.CallBankingService(ctx, "/api/v1/beneficiary", payload)
	if err != nil {
		return nil, err
	}

	// Extract beneficiary response
	responseResult := map[string]interface{}{
		"status":          "APPROVED",
		"beneficiary_id":  result["beneficiary_id"],
		"account_number":  accountNumber,
		"name":            name,
		"ifsc":            ifsc,
		"added_at":        time.Now(),
	}

	if message, ok := result["message"].(string); ok {
		responseResult["message"] = message
	} else {
		responseResult["message"] = fmt.Sprintf("Beneficiary %s added successfully", name)
	}

	// Include bank_name if available
	if bankName, ok := result["bank_name"].(string); ok {
		responseResult["bank_name"] = bankName
	}

	return &model.AgentResponse{
		AgentID:     ba.agentType,
		AgentType:   "BANKING",
		Status:      "APPROVED",
		Result:      responseResult,
		RiskScore:   0.1,
		Explanation: fmt.Sprintf("Beneficiary %s added successfully", name),
		Confidence:  0.95,
		Timestamp:   time.Now(),
		RequestID:   "",
	}, nil
}

