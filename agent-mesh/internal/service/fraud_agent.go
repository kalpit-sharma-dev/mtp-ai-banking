package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aibanking/agent-mesh/internal/model"
	"github.com/rs/zerolog/log"
)

// FraudAgent handles fraud detection using ML models and pattern analysis
type FraudAgent struct {
	*AgentBase
}

// NewFraudAgent creates a new fraud agent
func NewFraudAgent(base *AgentBase) *FraudAgent {
	return &FraudAgent{
		AgentBase: base,
	}
}

// Process processes a fraud check request
func (fa *FraudAgent) Process(ctx context.Context, req *model.AgentRequest) (*model.AgentResponse, error) {
	log.Info().
		Str("task", req.Task).
		Str("request_id", req.RequestID).
		Msg("Fraud agent processing request")

	inputCtx := req.InputContext
	data, ok := inputCtx["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data in input context")
	}

	// Extract transaction details
	amount, _ := data["amount"].(float64)
	toAccount, _ := data["to_account"].(string)
	userID, _ := inputCtx["user_id"].(string)

	// Perform fraud checks
	fraudScore := fa.calculateFraudScore(ctx, amount, toAccount, userID, inputCtx)
	
	// Determine status based on fraud score
	status := "APPROVED"
	explanation := "No fraud patterns detected"
	
	if fraudScore > 0.7 {
		status = "REJECTED"
		explanation = "High fraud risk detected. Transaction flagged for manual review."
	} else if fraudScore > 0.4 {
		status = "PENDING"
		explanation = "Moderate fraud risk. Additional verification required."
	}

	log.Info().
		Float64("fraud_score", fraudScore).
		Str("status", status).
		Msg("Fraud check completed")

	result := map[string]interface{}{
		"fraud_score":    fraudScore,
		"risk_level":     fa.getRiskLevel(fraudScore),
		"flags":          fa.getFraudFlags(ctx, amount, toAccount, userID, inputCtx),
		"recommendation": fa.getRecommendation(fraudScore),
	}

	return &model.AgentResponse{
		AgentID:     fa.agentType,
		AgentType:   "FRAUD",
		Status:      status,
		Result:      result,
		RiskScore:   fraudScore,
		Explanation: explanation,
		Confidence:  0.85,
		Timestamp:   time.Now(),
		RequestID:   req.RequestID,
	}, nil
}

// calculateFraudScore calculates fraud risk score using ML model simulation
func (fa *FraudAgent) calculateFraudScore(ctx context.Context, amount float64, toAccount string, userID string, context map[string]interface{}) float64 {
	score := 0.0

	// Amount-based risk
	if amount > 200000 {
		score += 0.4
	} else if amount > 100000 {
		score += 0.2
	} else if amount > 50000 {
		score += 0.1
	}

	// New beneficiary risk
	if beneficiaryAge, ok := context["beneficiary_age_days"].(float64); ok {
		if beneficiaryAge < 7 {
			score += 0.3 // New beneficiary (less than 7 days old)
		}
	} else {
		score += 0.2 // Unknown beneficiary age
	}

	// Time-based risk (unusual hours)
	if hour, ok := context["hour"].(float64); ok {
		if hour < 6 || hour > 23 {
			score += 0.15 // Unusual hours
		}
	}

	// Device anomaly
	if deviceRisk, ok := context["device_risk"].(float64); ok {
		score += deviceRisk * 0.2
	}

	// Location anomaly
	if locationRisk, ok := context["location_risk"].(float64); ok {
		score += locationRisk * 0.15
	}

	// Velocity check (too many transactions)
	if txnCount, ok := context["transaction_count_24h"].(float64); ok {
		if txnCount > 10 {
			score += 0.25
		} else if txnCount > 5 {
			score += 0.1
		}
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// getRiskLevel returns risk level based on score
func (fa *FraudAgent) getRiskLevel(score float64) string {
	if score > 0.7 {
		return "HIGH"
	} else if score > 0.4 {
		return "MEDIUM"
	}
	return "LOW"
}

// getFraudFlags returns list of fraud flags
func (fa *FraudAgent) getFraudFlags(ctx context.Context, amount float64, toAccount string, userID string, context map[string]interface{}) []string {
	flags := []string{}

	if amount > 100000 {
		flags = append(flags, "HIGH_AMOUNT")
	}

	if beneficiaryAge, ok := context["beneficiary_age_days"].(float64); ok && beneficiaryAge < 7 {
		flags = append(flags, "NEW_BENEFICIARY")
	}

	if txnCount, ok := context["transaction_count_24h"].(float64); ok && txnCount > 5 {
		flags = append(flags, "HIGH_VELOCITY")
	}

	if deviceRisk, ok := context["device_risk"].(float64); ok && deviceRisk > 0.5 {
		flags = append(flags, "DEVICE_ANOMALY")
	}

	return flags
}

// getRecommendation returns recommendation based on fraud score
func (fa *FraudAgent) getRecommendation(score float64) string {
	if score > 0.7 {
		return "BLOCK_TRANSACTION"
	} else if score > 0.4 {
		return "STEP_UP_AUTH"
	}
	return "PROCEED"
}

