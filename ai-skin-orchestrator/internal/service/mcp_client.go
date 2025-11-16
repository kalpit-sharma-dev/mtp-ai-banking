package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aibanking/ai-skin-orchestrator/internal/config"
	"github.com/aibanking/ai-skin-orchestrator/internal/model"
)

// MCPClient handles communication with the MCP Server (Layer 1)
type MCPClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewMCPClient creates a new MCP client
func NewMCPClient(cfg *config.MCPServerConfig) *MCPClient {
	return &MCPClient{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// SubmitTask submits a task to the MCP server
func (mc *MCPClient) SubmitTask(ctx context.Context, req *model.UserRequest, intent model.Intent, enrichedContext *model.EnrichedContext) (*model.AgentResponse, error) {
	// Prepare task request for MCP server
	taskReq := map[string]interface{}{
		"user_id":  req.UserID,
		"channel":  req.Channel,
		"intent":   string(intent.Type),
		"data":     intent.Entities,
		"context":  enrichedContext.Metadata,
	}

	if req.SessionID != "" {
		taskReq["session_id"] = req.SessionID
	}

	url := fmt.Sprintf("%s/api/v1/submit-task", mc.baseURL)
	
	body, err := json.Marshal(taskReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", mc.apiKey)

	resp, err := mc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to submit task: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("MCP server error: %s", string(respBody))
	}

	var taskResp struct {
		TaskID    string `json:"task_id"`
		SessionID string `json:"session_id"`
		Status    string `json:"status"`
	}

	if err := json.Unmarshal(respBody, &taskResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Wait for task to complete with retries
	maxRetries := 10
	retryDelay := 500 * time.Millisecond
	
	for i := 0; i < maxRetries; i++ {
		time.Sleep(retryDelay)
		result, err := mc.GetTaskResult(ctx, taskResp.TaskID)
		if err == nil {
			// Check if task is completed
			if result.Status == "APPROVED" || result.Status == "REJECTED" || result.Status == "COMPLETED" {
				return result, nil
			}
			// If still processing, wait and retry
		}
		// If error or still processing, retry
	}
	
	// Final attempt
	return mc.GetTaskResult(ctx, taskResp.TaskID)
}

// GetTaskResult retrieves task result from MCP server
func (mc *MCPClient) GetTaskResult(ctx context.Context, taskID string) (*model.AgentResponse, error) {
	url := fmt.Sprintf("%s/api/v1/get-result/%s", mc.baseURL, taskID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("X-API-Key", mc.apiKey)

	resp, err := mc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get result: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MCP server error: %s", string(respBody))
	}

	var result struct {
		TaskID      string                 `json:"task_id"`
		Status      string                 `json:"status"`
		Result      map[string]interface{} `json:"result"`
		RiskScore   float64                `json:"risk_score"`
		Explanation string                 `json:"explanation"`
		Error       string                 `json:"error,omitempty"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &model.AgentResponse{
		AgentID:     "mcp-agent",
		AgentType:   "ORCHESTRATED",
		Status:      result.Status,
		Result:      result.Result,
		RiskScore:   result.RiskScore,
		Explanation: result.Explanation,
		Confidence:  0.9,
		Timestamp:   time.Now(),
	}, nil
}

