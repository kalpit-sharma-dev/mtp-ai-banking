package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aibanking/banking-integrations/internal/model"
	"github.com/aibanking/banking-integrations/internal/service"
	"github.com/rs/zerolog/log"
)

// BankingController handles banking API requests
type BankingController struct {
	gateway *service.BankingGateway
}

// NewBankingController creates a new banking controller
func NewBankingController(gateway *service.BankingGateway) *BankingController {
	return &BankingController{
		gateway: gateway,
	}
}

// GetBalance handles GET /balance
func (bc *BankingController) GetBalance(w http.ResponseWriter, r *http.Request) {
	var req model.BalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	if req.UserID == "" || req.AccountID == "" {
		respondWithError(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	response, err := bc.gateway.GetBalance(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get balance", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// TransferFunds handles POST /transfer
func (bc *BankingController) TransferFunds(w http.ResponseWriter, r *http.Request) {
	var req model.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	if req.UserID == "" || req.FromAccount == "" || req.ToAccount == "" || req.Amount <= 0 {
		respondWithError(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	response, err := bc.gateway.TransferFunds(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to transfer funds", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetStatement handles POST /statement
func (bc *BankingController) GetStatement(w http.ResponseWriter, r *http.Request) {
	var req model.StatementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	if req.AccountID == "" || req.UserID == "" {
		respondWithError(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	// Set default dates if not provided
	if req.StartDate.IsZero() {
		req.StartDate = time.Now().AddDate(0, 0, -30)
	}
	if req.EndDate.IsZero() {
		req.EndDate = time.Now()
	}

	response, err := bc.gateway.GetStatement(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get statement", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// AddBeneficiary handles POST /beneficiary
func (bc *BankingController) AddBeneficiary(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID       string          `json:"user_id"`
		AccountNumber string         `json:"account_number"`
		IFSC         string          `json:"ifsc"`
		Name         string          `json:"name"`
		Channel      model.Channel   `json:"channel"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	if req.UserID == "" || req.AccountNumber == "" || req.IFSC == "" || req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	response, err := bc.gateway.AddBeneficiary(r.Context(), req.Channel, req.UserID, req.AccountNumber, req.IFSC, req.Name)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to add beneficiary", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// QueryDWH handles POST /dwh/query
func (bc *BankingController) QueryDWH(w http.ResponseWriter, r *http.Request) {
	var req model.DWHQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	if req.QueryType == "" {
		respondWithError(w, http.StatusBadRequest, "Query type is required", nil)
		return
	}

	response, err := bc.gateway.QueryDWH(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to query DWH", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetTransactionHistory handles GET /dwh/history/{userID}
func (bc *BankingController) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	days := 90 // Default 90 days
	if daysParam := r.URL.Query().Get("days"); daysParam != "" {
		// Parse days parameter if provided
		// For simplicity, using default
	}

	transactions, err := bc.gateway.GetTransactionHistory(r.Context(), userID, days)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get transaction history", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":     userID,
		"transactions": transactions,
		"count":       len(transactions),
	})
}

// HealthCheck handles GET /health
func (bc *BankingController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"banking-integrations"}`))
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, code int, message string, err error) {
	log.Error().Err(err).Str("message", message).Msg("Request error")
	
	response := map[string]interface{}{
		"error":   message,
		"code":    code,
		"details": "",
	}
	
	if err != nil {
		response["details"] = err.Error()
	}

	respondWithJSON(w, code, response)
}

