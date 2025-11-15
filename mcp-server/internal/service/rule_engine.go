package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aibanking/mcp-server/internal/model"
	"github.com/rs/zerolog/log"
)

// RuleEngine evaluates routing rules and business logic
type RuleEngine struct {
	rules map[string]interface{}
	mu    sync.RWMutex
}

// NewRuleEngine creates a new rule engine instance
func NewRuleEngine() *RuleEngine {
	engine := &RuleEngine{
		rules: make(map[string]interface{}),
	}
	
	// Load default rules
	engine.loadDefaultRules()
	
	return engine
}

// LoadRulesFromFile loads routing rules from a JSON/YAML file
func (re *RuleEngine) LoadRulesFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read rules file: %w", err)
	}

	var rules map[string]interface{}
	if err := json.Unmarshal(data, &rules); err != nil {
		return fmt.Errorf("failed to parse rules file: %w", err)
	}

	re.mu.Lock()
	re.rules = rules
	re.mu.Unlock()

	log.Info().Str("file", filePath).Msg("Rules loaded from file")
	return nil
}

// UploadRules uploads routing rules from a map
func (re *RuleEngine) UploadRules(rules map[string]interface{}) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	// Merge with existing rules
	for k, v := range rules {
		re.rules[k] = v
	}

	log.Info().Int("rule_count", len(rules)).Msg("Rules uploaded")
	return nil
}

// EvaluateRoutingRules evaluates rules to determine agent routing
func (re *RuleEngine) EvaluateRoutingRules(ctx context.Context, enrichedContext *model.Context, task *model.Task) (*model.RoutingDecision, error) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	// Check for intent-specific rules
	intentKey := fmt.Sprintf("intent:%s", task.Intent)
	if rule, exists := re.rules[intentKey]; exists {
		return re.applyRule(rule, enrichedContext, task)
	}

	// Check for channel-specific rules
	channelKey := fmt.Sprintf("channel:%s", task.Channel)
	if rule, exists := re.rules[channelKey]; exists {
		return re.applyRule(rule, enrichedContext, task)
	}

	// Check for risk-level rules
	riskKey := fmt.Sprintf("risk:%s", enrichedContext.RiskLevel)
	if rule, exists := re.rules[riskKey]; exists {
		return re.applyRule(rule, enrichedContext, task)
	}

	// No matching rule found
	return &model.RoutingDecision{
		SelectedAgentID: "",
		AgentType:       "",
		Confidence:      0.0,
		Reason:          "No matching routing rule",
		Context:         enrichedContext,
	}, nil
}

// applyRule applies a routing rule and returns a decision
func (re *RuleEngine) applyRule(rule interface{}, enrichedContext *model.Context, task *model.Task) (*model.RoutingDecision, error) {
	ruleMap, ok := rule.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid rule format")
	}

	agentType, ok := ruleMap["agent_type"].(string)
	if !ok {
		return nil, fmt.Errorf("rule missing agent_type")
	}

	reason, _ := ruleMap["reason"].(string)
	confidence := 0.9
	if conf, ok := ruleMap["confidence"].(float64); ok {
		confidence = conf
	}

	return &model.RoutingDecision{
		SelectedAgentID: "", // Will be resolved by router
		AgentType:       agentType,
		Confidence:      confidence,
		Reason:          reason,
		Context:         enrichedContext,
	}, nil
}

// loadDefaultRules loads default routing rules
func (re *RuleEngine) loadDefaultRules() {
	defaultRules := map[string]interface{}{
		"intent:TRANSFER_NEFT": map[string]interface{}{
			"agent_type": "GUARDRAIL",
			"reason":     "NEFT transfers require guardrail validation",
			"confidence": 0.9,
		},
		"intent:TRANSFER_RTGS": map[string]interface{}{
			"agent_type": "GUARDRAIL",
			"reason":     "RTGS transfers require guardrail validation",
			"confidence": 0.9,
		},
		"risk:HIGH": map[string]interface{}{
			"agent_type": "FRAUD",
			"reason":     "High-risk transaction requires fraud check",
			"confidence": 0.95,
		},
		"intent:APPLY_LOAN": map[string]interface{}{
			"agent_type": "CLEARANCE",
			"reason":     "Loan applications require clearance agent",
			"confidence": 0.9,
		},
	}

	re.mu.Lock()
	re.rules = defaultRules
	re.mu.Unlock()

	log.Info().Msg("Default rules loaded")
}

// GetRules returns all current rules
func (re *RuleEngine) GetRules() map[string]interface{} {
	re.mu.RLock()
	defer re.mu.RUnlock()

	rulesCopy := make(map[string]interface{})
	for k, v := range re.rules {
		rulesCopy[k] = v
	}

	return rulesCopy
}

// SaveRulesToFile saves current rules to a file
func (re *RuleEngine) SaveRulesToFile(filePath string) error {
	rules := re.GetRules()

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write rules file: %w", err)
	}

	log.Info().Str("file", filePath).Msg("Rules saved to file")
	return nil
}

