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
	agentType                  string
	agentName                  string
	endpoint                   string
	mcpBaseURL                 string
	mcpAPIKey                  string
	httpClient                 *http.Client
	mlModelsURL                string
	mlModelsKey                string
	mlModelsEnabled            bool
	bankingIntegrationsURL     string
	bankingIntegrationsKey     string
	bankingIntegrationsEnabled bool
}

// NewAgentBase creates a new agent base
func NewAgentBase(agentType, agentName, endpoint string, mcpConfig *config.MCPServerConfig, mlConfig *config.MLModelsConfig, bankingConfig *config.BankingIntegrationsConfig) *AgentBase {
	return &AgentBase{
		agentType:                  agentType,
		agentName:                  agentName,
		endpoint:                   endpoint,
		mcpBaseURL:                 mcpConfig.BaseURL,
		mcpAPIKey:                  mcpConfig.APIKey,
		mlModelsURL:                mlConfig.BaseURL,
		mlModelsKey:                mlConfig.APIKey,
		mlModelsEnabled:            mlConfig.Enabled,
		bankingIntegrationsURL:     bankingConfig.BaseURL,
		bankingIntegrationsKey:     bankingConfig.APIKey,
		bankingIntegrationsEnabled: bankingConfig.Enabled,
		httpClient: &http.Client{
			Timeout: time.Duration(mcpConfig.Timeout) * time.Second,
		},
	}
}

// RegisterWithMCP registers this agent with the MCP Server with retry logic
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

	// Retry logic: try up to 3 times with increasing delays
	maxRetries := 3
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(bytes.NewBuffer(body)))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("X-API-Key", ab.mcpAPIKey)

		resp, err := ab.httpClient.Do(httpReq)
		if err != nil {
			if attempt < maxRetries {
				log.Warn().
					Int("attempt", attempt).
					Int("max_retries", maxRetries).
					Err(err).
					Msg("Failed to register agent, retrying...")
				time.Sleep(retryDelay)
				retryDelay *= 2 // Exponential backoff
				continue
			}
			return fmt.Errorf("failed to register agent after %d attempts: %w", maxRetries, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			respBody, _ := io.ReadAll(resp.Body)
			if attempt < maxRetries {
				log.Warn().
					Int("attempt", attempt).
					Int("status_code", resp.StatusCode).
					Str("response", string(respBody)).
					Msg("Registration failed, retrying...")
				time.Sleep(retryDelay)
				retryDelay *= 2
				continue
			}
			return fmt.Errorf("registration failed: %s", string(respBody))
		}

		log.Info().
			Str("agent_type", ab.agentType).
			Str("agent_name", ab.agentName).
			Msg("Agent registered with MCP Server")

		return nil
	}

	return fmt.Errorf("failed to register agent after %d attempts", maxRetries)
}

// ProcessRequest is the interface that all agents must implement
type ProcessRequest interface {
	Process(ctx context.Context, req *model.AgentRequest) (*model.AgentResponse, error)
}

// CallMLService calls the ML Models service
func (ab *AgentBase) CallMLService(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	if !ab.mlModelsEnabled {
		return nil, fmt.Errorf("ML models service is disabled")
	}

	url := fmt.Sprintf("%s%s", ab.mlModelsURL, endpoint)
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if ab.mlModelsKey != "" {
		httpReq.Header.Set("X-API-Key", ab.mlModelsKey)
	}

	resp, err := ab.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service error: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// CallBankingService calls the Banking Integrations service
func (ab *AgentBase) CallBankingService(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	if !ab.bankingIntegrationsEnabled {
		return nil, fmt.Errorf("banking integrations service is disabled")
	}

	url := fmt.Sprintf("%s%s", ab.bankingIntegrationsURL, endpoint)
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if ab.bankingIntegrationsKey != "" {
		httpReq.Header.Set("X-API-Key", ab.bankingIntegrationsKey)
	}

	log.Info().
		Str("url", url).
		Interface("payload", payload).
		Str("banking_integrations_url", ab.bankingIntegrationsURL).
		Bool("enabled", ab.bankingIntegrationsEnabled).
		Msg("Calling Banking Integrations service")
	
	resp, err := ab.httpClient.Do(httpReq)
	if err != nil {
		log.Error().
			Err(err).
			Str("url", url).
			Str("full_url", url).
			Str("base_url", ab.bankingIntegrationsURL).
			Msg("❌ CRITICAL: Failed to call Banking Integrations service - check if service is running on port 7000")
		return nil, fmt.Errorf("failed to call banking service: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().
			Err(err).
			Str("url", url).
			Msg("Failed to read Banking Integrations response")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Accept both 200 (OK) and 201 (Created) as success status codes
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("url", url).
			Str("response", string(respBody)).
			Msg("❌ Banking Integrations service returned error status")
		return nil, fmt.Errorf("banking service error (status %d): %s", resp.StatusCode, string(respBody))
	}
	
	responsePreview := string(respBody)
	if len(responsePreview) > 100 {
		responsePreview = responsePreview[:100] + "..."
	}
	log.Info().
		Str("url", url).
		Int("status_code", resp.StatusCode).
		Str("response_preview", responsePreview).
		Msg("✅ Banking Integrations service call successful")

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}
