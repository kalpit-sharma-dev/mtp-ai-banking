package controller

import (
	"encoding/json"
	"net/http"

	"github.com/aibanking/mcp-server/internal/model"
	"github.com/aibanking/mcp-server/internal/service"
)

// SessionController handles session-related HTTP requests
type SessionController struct {
	sessionManager *service.SessionManager
}

// NewSessionController creates a new session controller
func NewSessionController(sessionManager *service.SessionManager) *SessionController {
	return &SessionController{
		sessionManager: sessionManager,
	}
}

// GetSession handles GET /get-session/{sessionID}
func (sc *SessionController) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")
	if sessionID == "" {
		RespondWithError(w, http.StatusBadRequest, "Session ID is required", nil)
		return
	}

	session, err := sc.sessionManager.GetSession(r.Context(), sessionID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Session not found", err)
		return
	}

	response := &model.SessionResponse{
		SessionID:   session.SessionID,
		UserID:      session.UserID,
		Channel:     session.Channel,
		Context:     session.Context,
		TaskHistory: session.TaskHistory,
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// CreateSession handles POST /create-session
func (sc *SessionController) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req model.SessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Channel == "" {
		RespondWithError(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	session, err := sc.sessionManager.CreateSession(r.Context(), &req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create session", err)
		return
	}

	response := &model.SessionResponse{
		SessionID:   session.SessionID,
		UserID:      session.UserID,
		Channel:     session.Channel,
		Context:     session.Context,
		TaskHistory: session.TaskHistory,
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
	}

	RespondWithJSON(w, http.StatusCreated, response)
}

