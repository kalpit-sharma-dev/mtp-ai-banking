package controller

import (
	"encoding/json"
	"net/http"

	"github.com/aibanking/ai-skin-orchestrator/internal/model"
	"github.com/aibanking/ai-skin-orchestrator/internal/service"
	"github.com/rs/zerolog/log"
)

// OrchestratorController handles orchestration requests
type OrchestratorController struct {
	orchestrator  *service.Orchestrator
	sessionService *service.SessionService
}

// NewOrchestratorController creates a new orchestrator controller
func NewOrchestratorController(orchestrator *service.Orchestrator, sessionService *service.SessionService) *OrchestratorController {
	return &OrchestratorController{
		orchestrator:   orchestrator,
		sessionService: sessionService,
	}
}

// ProcessRequest handles POST /process
func (oc *OrchestratorController) ProcessRequest(w http.ResponseWriter, r *http.Request) {
	var req model.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Channel == "" || req.Input == "" {
		respondWithError(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	// Set default input type if not provided
	if req.InputType == "" {
		req.InputType = "natural_language"
	}

	// Get or create session
	session := oc.sessionService.GetOrCreateSession(r.Context(), req.SessionID, req.UserID, req.Channel)
	req.SessionID = session.SessionID

	// Add user message to session
	oc.sessionService.AddMessage(session.SessionID, "user", req.Input)

	// Process request
	response, err := oc.orchestrator.ProcessRequest(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to process request", err)
		return
	}

	// Add bot response to session
	if response.Explanation != "" {
		oc.sessionService.AddMessage(session.SessionID, "bot", response.Explanation)
	}

	// Add session ID to response
	if response.FinalResult == nil {
		response.FinalResult = make(map[string]interface{})
	}
	response.FinalResult["session_id"] = session.SessionID

	respondWithJSON(w, http.StatusOK, response)
}

// HealthCheck handles GET /health
func (oc *OrchestratorController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"ai-skin-orchestrator"}`))
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

