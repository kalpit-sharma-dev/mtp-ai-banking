package controller

import (
	"encoding/json"
	"fmt"
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
	// Parse request with flexible date handling
	var reqData struct {
		AccountID string `json:"account_id"`
		UserID    string `json:"user_id"`
		StartDate string `json:"start_date"` // Accept as string to handle multiple formats
		EndDate   string `json:"end_date"`   // Accept as string to handle multiple formats
		Channel   model.Channel `json:"channel"`
		Limit     int    `json:"limit,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	if reqData.AccountID == "" || reqData.UserID == "" {
		respondWithError(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	// Parse dates with multiple format support
	var startDate, endDate time.Time
	var err error
	
	// Parse start_date
	if reqData.StartDate != "" {
		startDate, err = parseDate(reqData.StartDate)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid start_date format: %s", err.Error()), nil)
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30) // Default to 30 days ago
	}
	
	// Parse end_date
	if reqData.EndDate != "" {
		endDate, err = parseDate(reqData.EndDate)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid end_date format: %s", err.Error()), nil)
			return
		}
	} else {
		endDate = time.Now() // Default to now
	}

	// Create StatementRequest with parsed dates
	req := model.StatementRequest{
		AccountID: reqData.AccountID,
		UserID:    reqData.UserID,
		StartDate: startDate,
		EndDate:   endDate,
		Channel:   reqData.Channel,
		Limit:     reqData.Limit,
	}

	response, err := bc.gateway.GetStatement(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get statement", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// parseDate parses date strings in multiple formats
func parseDate(dateStr string) (time.Time, error) {
	// Try RFC3339 format first (full timestamp)
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}
	
	// Try RFC3339Nano format
	if t, err := time.Parse(time.RFC3339Nano, dateStr); err == nil {
		return t, nil
	}
	
	// Try date-only format (YYYY-MM-DD) - set to start of day in local timezone
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		// Set to start of day (00:00:00)
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local), nil
	}
	
	// Try date with time format (YYYY-MM-DD HH:MM:SS)
	if t, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
		return t, nil
	}
	
	// Try date with time and timezone (YYYY-MM-DD HH:MM:SS +0000)
	if t, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr); err == nil {
		return t, nil
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s (supported formats: RFC3339, YYYY-MM-DD, YYYY-MM-DD HH:MM:SS)", dateStr)
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

