package controller

import (
	"encoding/json"
	"net/http"

	"github.com/aibanking/agent-mesh/internal/model"
	"github.com/aibanking/agent-mesh/internal/service"
	"github.com/rs/zerolog/log"
)

// AgentController handles agent requests
type AgentController struct {
	agentProcessor service.ProcessRequest
	agentType      string
}

// NewAgentController creates a new agent controller
func NewAgentController(agentProcessor service.ProcessRequest, agentType string) *AgentController {
	return &AgentController{
		agentProcessor: agentProcessor,
		agentType:     agentType,
	}
}

// ProcessRequest handles POST /process
func (ac *AgentController) ProcessRequest(w http.ResponseWriter, r *http.Request) {
	var req model.AgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	// Set agent type if not provided
	if req.AgentID == "" {
		req.AgentID = ac.agentType
	}

	// Process request
	response, err := ac.agentProcessor.Process(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to process request", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

// HealthCheck handles GET /health
func (ac *AgentController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","agent_type":"` + ac.agentType + `"}`))
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

