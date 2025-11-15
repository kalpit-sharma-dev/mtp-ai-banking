package controller

import (
	"encoding/json"
	"net/http"

	"github.com/aibanking/mcp-server/internal/model"
	"github.com/aibanking/mcp-server/internal/service"
)

// AgentController handles agent-related HTTP requests
type AgentController struct {
	agentRegistry *service.AgentRegistry
}

// NewAgentController creates a new agent controller
func NewAgentController(agentRegistry *service.AgentRegistry) *AgentController {
	return &AgentController{
		agentRegistry: agentRegistry,
	}
}

// RegisterAgent handles POST /register-agent
func (ac *AgentController) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	var req model.AgentRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	// Validate required fields
	if req.Name == "" || req.Type == "" || req.Endpoint == "" || len(req.Capabilities) == 0 {
		RespondWithError(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	// Register agent
	agent, err := ac.agentRegistry.RegisterAgent(r.Context(), &req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to register agent", err)
		return
	}

	response := &model.AgentResponse{
		AgentID:      agent.AgentID,
		Name:         agent.Name,
		Type:         string(agent.Type),
		Status:       string(agent.Status),
		RegisteredAt: agent.RegisteredAt,
		Message:      "Agent registered successfully",
	}

	RespondWithJSON(w, http.StatusCreated, response)
}

// GetAgent handles GET /agent/{agentID}
func (ac *AgentController) GetAgent(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentID")
	if agentID == "" {
		RespondWithError(w, http.StatusBadRequest, "Agent ID is required", nil)
		return
	}

	agent, err := ac.agentRegistry.GetAgent(r.Context(), agentID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Agent not found", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, agent)
}

// GetAllAgents handles GET /agents
func (ac *AgentController) GetAllAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := ac.agentRegistry.GetAllAgents(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get agents", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	})
}
