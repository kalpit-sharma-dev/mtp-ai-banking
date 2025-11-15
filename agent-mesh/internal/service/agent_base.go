package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aibanking/agent-mesh/internal/config"
	"github.com/aibanking/agent-mesh/internal/model"
	"github.com/rs/zerolog/log"
)

// AgentBase provides base functionality for all agents
type AgentBase struct {
	agentType   string
	agentName   string
	endpoint    string
	mcpBaseURL  string
	mcpAPIKey   string
	httpClient  *http.Client
}

// NewAgentBase creates a new agent base
func NewAgentBase(agentType, agentName, endpoint string, mcpConfig *config.MCPServerConfig) *AgentBase {
	return &AgentBase{
		agentType:  agentType,
		agentName:  agentName,
		endpoint:   endpoint,
		mcpBaseURL: mcpConfig.BaseURL,
		mcpAPIKey: mcpConfig.APIKey,
		httpClient: &http.Client{
			Timeout: time.Duration(mcpConfig.Timeout) * time.Second,
		},
	}
}

// RegisterWithMCP registers this agent with the MCP Server
func (ab *AgentBase) RegisterWithMCP(ctx context.Context, capabilities []string) error {
	req := map[string]interface{}{
		"name":         ab.agentName,
		"type":         ab.agentType,
		"endpoint":     ab.endpoint,
		"capabilities": capabilities,
		"metadata": map[string]interface{}{
			"registered_at": time.Now(),
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/register-agent", ab.mcpBaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", ab.mcpAPIKey)

	resp, err := ab.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %s", string(respBody))
	}

	log.Info().
		Str("agent_type", ab.agentType).
		Str("agent_name", ab.agentName).
		Msg("Agent registered with MCP Server")

	return nil
}

// ProcessRequest is the interface that all agents must implement
type ProcessRequest interface {
	Process(ctx context.Context, req *model.AgentRequest) (*model.AgentResponse, error)
}

