package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aibanking/agent-mesh/internal/model"
	"github.com/rs/zerolog/log"
)

// ScoringAgent handles credit scoring, fraud scoring, and risk assessment
type ScoringAgent struct {
	*AgentBase
}

// NewScoringAgent creates a new scoring agent
func NewScoringAgent(base *AgentBase) *ScoringAgent {
	return &ScoringAgent{
		AgentBase: base,
	}
}

// Process processes a scoring request
func (sa *ScoringAgent) Process(ctx context.Context, req *model.AgentRequest) (*model.AgentResponse, error) {
	log.Info().
		Str("task", req.Task).
		Str("request_id", req.RequestID).
		Msg("Scoring agent processing request")

	inputCtx := req.InputContext
	scoreType, _ := inputCtx["score_type"].(string)
	if scoreType == "" {
		scoreType = "CREDIT" // Default
	}

	var result map[string]interface{}
	var riskScore float64
	var explanation string

	switch scoreType {
	case "CREDIT":
		result, riskScore, explanation = sa.calculateCreditScore(ctx, inputCtx)
	case "FRAUD":
		result, riskScore, explanation = sa.calculateFraudScore(ctx, inputCtx)
	case "RISK":
		result, riskScore, explanation = sa.calculateRiskScore(ctx, inputCtx)
	default:
		return nil, fmt.Errorf("unsupported score type: %s", scoreType)
	}

	log.Info().
		Str("score_type", scoreType).
		Float64("risk_score", riskScore).
		Msg("Scoring completed")

	return &model.AgentResponse{
		AgentID:     sa.agentType,
		AgentType:   "SCORING",
		Status:      "APPROVED",
		Result:      result,
		RiskScore:   riskScore,
		Explanation: explanation,
		Confidence:  0.9,
		Timestamp:   time.Now(),
		RequestID:   req.RequestID,
	}, nil
}

// calculateCreditScore calculates credit score
func (sa *ScoringAgent) calculateCreditScore(ctx context.Context, context map[string]interface{}) (map[string]interface{}, float64, string) {
	// Extract user profile data
	accountAge, _ := context["account_age_days"].(float64)
	income, _ := context["income"].(float64)
	delinquency, _ := context["delinquency_count"].(float64)
	loanHistory, _ := context["loan_history_count"].(float64)
	balance, _ := context["balance"].(float64)

	// Base score
	score := 600.0

	// Account age factor (max +50 points)
	if accountAge > 365 {
		score += 50
	} else if accountAge > 180 {
		score += 30
	} else if accountAge > 90 {
		score += 15
	}

	// Income factor (max +100 points)
	if income > 100000 {
		score += 100
	} else if income > 50000 {
		score += 60
	} else if income > 25000 {
		score += 30
	}

	// Delinquency penalty (max -100 points)
	score -= delinquency * 20
	if score < 300 {
		score = 300
	}

	// Loan history bonus (max +50 points)
	if loanHistory > 0 {
		score += 30
	}

	// Balance factor (max +50 points)
	if balance > 100000 {
		score += 50
	} else if balance > 50000 {
		score += 30
	} else if balance > 10000 {
		score += 15
	}

	// Cap at 850
	if score > 850 {
		score = 850
	}

	// Convert to risk score (inverse: higher credit score = lower risk)
	riskScore := 1.0 - (score / 850.0)

	result := map[string]interface{}{
		"credit_score":  int(score),
		"score_range":  sa.getScoreRange(score),
		"risk_category": sa.getRiskCategory(riskScore),
		"factors": map[string]interface{}{
			"account_age":    accountAge,
			"income":         income,
			"delinquency":    delinquency,
			"loan_history":   loanHistory,
			"balance":        balance,
		},
	}

	explanation := fmt.Sprintf("Credit score calculated: %d (%s)", int(score), sa.getScoreRange(score))

	return result, riskScore, explanation
}

// calculateFraudScore calculates fraud risk score
func (sa *ScoringAgent) calculateFraudScore(ctx context.Context, context map[string]interface{}) (map[string]interface{}, float64, string) {
	amount, _ := context["amount"].(float64)
	deviceRisk, _ := context["device_risk"].(float64)
	locationRisk, _ := context["location_risk"].(float64)
	velocity, _ := context["transaction_count_24h"].(float64)

	score := 0.0

	// Amount factor
	if amount > 200000 {
		score += 0.4
	} else if amount > 100000 {
		score += 0.2
	}

	// Device risk
	score += deviceRisk * 0.3

	// Location risk
	score += locationRisk * 0.2

	// Velocity factor
	if velocity > 10 {
		score += 0.3
	} else if velocity > 5 {
		score += 0.15
	}

	if score > 1.0 {
		score = 1.0
	}

	result := map[string]interface{}{
		"fraud_score":   score,
		"risk_level":    sa.getRiskLevel(score),
		"recommendation": sa.getFraudRecommendation(score),
	}

	explanation := fmt.Sprintf("Fraud risk score: %.2f (%s)", score, sa.getRiskLevel(score))

	return result, score, explanation
}

// calculateRiskScore calculates overall risk score
func (sa *ScoringAgent) calculateRiskScore(ctx context.Context, context map[string]interface{}) (map[string]interface{}, float64, string) {
	creditScore, _ := context["credit_score"].(float64)
	fraudScore, _ := context["fraud_score"].(float64)
	amount, _ := context["amount"].(float64)

	// Normalize credit score to 0-1 (inverse)
	creditRisk := 1.0 - (creditScore / 850.0)

	// Weighted combination
	overallRisk := (creditRisk * 0.4) + (fraudScore * 0.4) + (sa.getAmountRisk(amount) * 0.2)

	if overallRisk > 1.0 {
		overallRisk = 1.0
	}

	result := map[string]interface{}{
		"overall_risk":  overallRisk,
		"risk_category": sa.getRiskCategory(overallRisk),
		"components": map[string]interface{}{
			"credit_risk": creditRisk,
			"fraud_risk":  fraudScore,
			"amount_risk": sa.getAmountRisk(amount),
		},
	}

	explanation := fmt.Sprintf("Overall risk score: %.2f (%s)", overallRisk, sa.getRiskCategory(overallRisk))

	return result, overallRisk, explanation
}

// Helper functions
func (sa *ScoringAgent) getScoreRange(score float64) string {
	if score >= 750 {
		return "EXCELLENT"
	} else if score >= 700 {
		return "GOOD"
	} else if score >= 650 {
		return "FAIR"
	} else if score >= 600 {
		return "POOR"
	}
	return "VERY_POOR"
}

func (sa *ScoringAgent) getRiskCategory(risk float64) string {
	if risk < 0.3 {
		return "LOW"
	} else if risk < 0.6 {
		return "MEDIUM"
	}
	return "HIGH"
}

func (sa *ScoringAgent) getRiskLevel(score float64) string {
	if score > 0.7 {
		return "HIGH"
	} else if score > 0.4 {
		return "MEDIUM"
	}
	return "LOW"
}

func (sa *ScoringAgent) getFraudRecommendation(score float64) string {
	if score > 0.7 {
		return "BLOCK"
	} else if score > 0.4 {
		return "REVIEW"
	}
	return "APPROVE"
}

func (sa *ScoringAgent) getAmountRisk(amount float64) float64 {
	if amount > 200000 {
		return 0.8
	} else if amount > 100000 {
		return 0.5
	} else if amount > 50000 {
		return 0.3
	}
	return 0.1
}

