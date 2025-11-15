package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aibanking/mcp-server/internal/model"
	"github.com/aibanking/mcp-server/internal/service"
	"github.com/gorilla/mux"
)

// TaskController handles task-related HTTP requests
type TaskController struct {
	orchestrator *service.Orchestrator
	taskManager  *service.TaskManager
}

// NewTaskController creates a new task controller
func NewTaskController(orchestrator *service.Orchestrator, taskManager *service.TaskManager) *TaskController {
	return &TaskController{
		orchestrator: orchestrator,
		taskManager:  taskManager,
	}
}

// SubmitTask handles POST /submit-task
func (tc *TaskController) SubmitTask(w http.ResponseWriter, r *http.Request) {
	var req model.TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Channel == "" || req.Intent == "" {
		RespondWithError(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	// Process task
	response, err := tc.orchestrator.ProcessTask(r.Context(), &req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to process task", err)
		return
	}

	RespondWithJSON(w, http.StatusAccepted, response)
}

// GetTaskResult handles GET /get-result/{taskID}
func (tc *TaskController) GetTaskResult(w http.ResponseWriter, r *http.Request) {
	// Get taskID from URL path using gorilla/mux
	vars := mux.Vars(r)
	taskID := vars["taskID"]
	
	// Fallback: try query parameter
	if taskID == "" {
		taskID = r.URL.Query().Get("taskID")
	}
	
	// Fallback: try to extract from path manually
	if taskID == "" {
		path := r.URL.Path
		// Path format: /api/v1/get-result/{taskID}
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "get-result" && i+1 < len(parts) {
				taskID = parts[i+1]
				break
			}
		}
	}

	if taskID == "" {
		RespondWithError(w, http.StatusBadRequest, "Task ID is required", nil)
		return
	}

	task, err := tc.taskManager.GetTask(r.Context(), taskID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Task not found", err)
		return
	}

	response := &model.TaskResultResponse{
		TaskID:      task.TaskID,
		Status:      string(task.Status),
		Result:      task.Result,
		RiskScore:   task.RiskScore,
		Explanation: task.Explanation,
		Error:       task.Error,
		CompletedAt: task.CompletedAt,
	}

	RespondWithJSON(w, http.StatusOK, response)
}

