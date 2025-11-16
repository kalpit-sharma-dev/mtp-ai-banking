package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aibanking/agent-mesh/internal/model"
	"github.com/rs/zerolog/log"
)

// GuardrailAgent handles RBI regulations and bank policy validation
type GuardrailAgent struct {
	*AgentBase
}

// NewGuardrailAgent creates a new guardrail agent
func NewGuardrailAgent(base *AgentBase) *GuardrailAgent {
	return &GuardrailAgent{
		AgentBase: base,
	}
}

// Process processes a guardrail validation request
func (ga *GuardrailAgent) Process(ctx context.Context, req *model.AgentRequest) (*model.AgentResponse, error) {
	log.Info().
		Str("task", req.Task).
		Str("request_id", req.RequestID).
		Msg("Guardrail agent processing request")

	inputCtx := req.InputContext
	data, ok := inputCtx["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data in input context")
	}

	amount, _ := data["amount"].(float64)
	userID, _ := inputCtx["user_id"].(string)

	// Perform guardrail checks
	checks := ga.performGuardrailChecks(ctx, amount, userID, inputCtx)
	
	// Determine if all checks passed
	allPassed := true
	failedChecks := []string{}
	
	for check, passed := range checks {
		if !passed {
			allPassed = false
			failedChecks = append(failedChecks, check)
		}
	}

	status := "APPROVED"
	explanation := "All guardrail checks passed"
	
	if !allPassed {
		status = "REJECTED"
		explanation = fmt.Sprintf("Guardrail checks failed: %v", failedChecks)
	}

	log.Info().
		Bool("all_passed", allPassed).
		Strs("failed_checks", failedChecks).
		Msg("Guardrail validation completed")

	result := map[string]interface{}{
		"checks":         checks,
		"all_passed":     allPassed,
		"failed_checks":  failedChecks,
		"validated_rules": ga.getValidatedRules(checks),
	}

	return &model.AgentResponse{
		AgentID:     ga.agentType,
		AgentType:   "GUARDRAIL",
		Status:      status,
		Result:      result,
		RiskScore:   ga.calculateRiskScore(checks),
		Explanation: explanation,
		Confidence:  0.95,
		Timestamp:   time.Now(),
		RequestID:   req.RequestID,
	}, nil
}

// performGuardrailChecks performs all guardrail validations
func (ga *GuardrailAgent) performGuardrailChecks(ctx context.Context, amount float64, userID string, context map[string]interface{}) map[string]bool {
	checks := make(map[string]bool)

	// Daily limit check (RBI regulation: 2 lakh for savings account)
	dailyLimit := 200000.0
	if dailyUsed, ok := context["daily_transaction_amount"].(float64); ok {
		checks["daily_limit"] = (dailyUsed + amount) <= dailyLimit
	} else {
		checks["daily_limit"] = amount <= dailyLimit
	}

	// Single transaction limit
	singleTxnLimit := 100000.0
	checks["single_transaction_limit"] = amount <= singleTxnLimit

	// Velocity check (max 10 transactions per day)
	if txnCount, ok := context["transaction_count_24h"].(float64); ok {
		checks["velocity_limit"] = txnCount < 10
	} else {
		checks["velocity_limit"] = true
	}

	// Beneficiary age check (minimum 24 hours for NEFT/RTGS)
	// For small amounts (< 10,000), allow transfers to new beneficiaries
	// For larger amounts, require beneficiary to be at least 1 day old
	if beneficiaryAge, ok := context["beneficiary_age_days"].(float64); ok {
		if amount < 10000.0 {
			// Allow small transfers to new beneficiaries
			checks["beneficiary_age"] = true
		} else {
			// For larger amounts, require at least 1 day old
			checks["beneficiary_age"] = beneficiaryAge >= 1
		}
	} else {
		// Unknown beneficiary age - allow for small amounts, reject for large amounts
		if amount < 10000.0 {
			checks["beneficiary_age"] = true // Allow small transfers
		} else {
			checks["beneficiary_age"] = false // Reject large transfers to unknown beneficiaries
		}
	}

	// KYC status check
	if kycStatus, ok := context["kyc_status"].(string); ok {
		checks["kyc_verified"] = kycStatus == "VERIFIED"
	} else {
		checks["kyc_verified"] = true // Assume verified if not provided
	}

	// Account status check
	if accountStatus, ok := context["account_status"].(string); ok {
		checks["account_active"] = accountStatus == "ACTIVE"
	} else {
		checks["account_active"] = true
	}

	// RBI blacklist check
	checks["rbi_blacklist"] = !ga.isBlacklisted(ctx, userID)

	return checks
}

// getValidatedRules returns list of validated rules
func (ga *GuardrailAgent) getValidatedRules(checks map[string]bool) []string {
	rules := []string{}
	for check, passed := range checks {
		if passed {
			rules = append(rules, check)
		}
	}
	return rules
}

// calculateRiskScore calculates risk score based on failed checks
func (ga *GuardrailAgent) calculateRiskScore(checks map[string]bool) float64 {
	failedCount := 0
	totalCount := len(checks)

	for _, passed := range checks {
		if !passed {
			failedCount++
		}
	}

	if totalCount == 0 {
		return 0.0
	}

	return float64(failedCount) / float64(totalCount)
}

// isBlacklisted checks if user/account is blacklisted
func (ga *GuardrailAgent) isBlacklisted(ctx context.Context, userID string) bool {
	// Mock implementation - in production would query RBI blacklist database
	return false
}

