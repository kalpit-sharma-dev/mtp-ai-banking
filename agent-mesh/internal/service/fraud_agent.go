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

// calculateFraudScore calculates fraud risk score using ML model
func (fa *FraudAgent) calculateFraudScore(ctx context.Context, amount float64, toAccount string, userID string, context map[string]interface{}) float64 {
	// Try to call ML Models service first
	if fa.mlModelsEnabled {
		score, err := fa.callMLFraudModel(ctx, amount, context)
		if err == nil {
			log.Info().Float64("fraud_score", score).Msg("Fraud score from ML model")
			return score
		}
		log.Warn().Err(err).Msg("ML model call failed, using fallback")
	}

	// Fallback to rule-based calculation
	return fa.calculateFraudScoreFallback(amount, toAccount, userID, context)
}

// callMLFraudModel calls the ML Models service for fraud prediction
func (fa *FraudAgent) callMLFraudModel(ctx context.Context, amount float64, context map[string]interface{}) (float64, error) {
	// Extract features from context
	hour := 12.0
	if h, ok := context["hour"].(float64); ok {
		hour = h
	}
	dayOfWeek := 3.0
	if d, ok := context["day_of_week"].(float64); ok {
		dayOfWeek = d
	}
	txnCount24h := 0.0
	if t, ok := context["transaction_count_24h"].(float64); ok {
		txnCount24h = t
	}
	txnCount7d := 0.0
	if t, ok := context["transaction_count_7d"].(float64); ok {
		txnCount7d = t
	}
	avgAmount7d := 10000.0
	if a, ok := context["avg_amount_7d"].(float64); ok {
		avgAmount7d = a
	}
	beneficiaryAge := 365.0
	if b, ok := context["beneficiary_age_days"].(float64); ok {
		beneficiaryAge = b
	}
	deviceRisk := 0.0
	if d, ok := context["device_risk"].(float64); ok {
		deviceRisk = d
	}
	locationRisk := 0.0
	if l, ok := context["location_risk"].(float64); ok {
		locationRisk = l
	}
	userAccountAge := 365.0
	if u, ok := context["user_account_age_days"].(float64); ok {
		userAccountAge = u
	}
	userBalance := 100000.0
	if u, ok := context["user_balance"].(float64); ok {
		userBalance = u
	}

	payload := map[string]interface{}{
		"amount":                  amount,
		"hour":                    hour,
		"day_of_week":             dayOfWeek,
		"transaction_count_24h":   txnCount24h,
		"transaction_count_7d":    txnCount7d,
		"avg_amount_7d":           avgAmount7d,
		"beneficiary_age_days":    beneficiaryAge,
		"device_risk":             deviceRisk,
		"location_risk":           locationRisk,
		"user_account_age_days":  userAccountAge,
		"user_balance":            userBalance,
		"is_new_beneficiary":     beneficiaryAge < 7,
		"is_unusual_hour":         hour < 6 || hour > 23,
		"amount_vs_avg_ratio":    amount / avgAmount7d,
		"velocity_score":          txnCount24h / 10.0,
	}

	result, err := fa.CallMLService(ctx, "/api/v1/fraud/predict", payload)
	if err != nil {
		return 0, err
	}

	// Extract fraud score from response
	if resultData, ok := result["result"].(map[string]interface{}); ok {
		if fraudScore, ok := resultData["fraud_score"].(float64); ok {
			return fraudScore, nil
		}
	}

	return 0, fmt.Errorf("invalid response format from ML service")
}

// calculateFraudScoreFallback uses rule-based calculation as fallback
func (fa *FraudAgent) calculateFraudScoreFallback(amount float64, toAccount string, userID string, context map[string]interface{}) float64 {
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
			score += 0.3
		}
	} else {
		score += 0.2
	}

	// Time-based risk (unusual hours)
	if hour, ok := context["hour"].(float64); ok {
		if hour < 6 || hour > 23 {
			score += 0.15
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

	// Velocity check
	if txnCount, ok := context["transaction_count_24h"].(float64); ok {
		if txnCount > 10 {
			score += 0.25
		} else if txnCount > 5 {
			score += 0.1
		}
	}

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

