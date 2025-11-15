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
	// Try to call ML Models service first
	if sa.mlModelsEnabled {
		result, riskScore, explanation, err := sa.callMLCreditModel(ctx, context)
		if err == nil {
			log.Info().Msg("Credit score from ML model")
			return result, riskScore, explanation
		}
		log.Warn().Err(err).Msg("ML model call failed, using fallback")
	}

	// Fallback to rule-based calculation
	return sa.calculateCreditScoreFallback(context)
}

// callMLCreditModel calls the ML Models service for credit scoring
func (sa *ScoringAgent) callMLCreditModel(ctx context.Context, context map[string]interface{}) (map[string]interface{}, float64, string, error) {
	// Extract features from context
	accountAge := 365.0
	if a, ok := context["account_age_days"].(float64); ok {
		accountAge = a
	}
	income := 50000.0
	if i, ok := context["income"].(float64); ok {
		income = i
	}
	balance := 100000.0
	if b, ok := context["balance"].(float64); ok {
		balance = b
	}
	txnCount30d := 10.0
	if t, ok := context["transaction_count_30d"].(float64); ok {
		txnCount30d = t
	}
	delinquency := 0.0
	if d, ok := context["delinquency_count"].(float64); ok {
		delinquency = d
	}
	loanHistory := 0.0
	if l, ok := context["loan_history_count"].(float64); ok {
		loanHistory = l
	}
	avgTxnAmount := 10000.0
	if a, ok := context["avg_transaction_amount"].(float64); ok {
		avgTxnAmount = a
	}
	creditUtilization := 0.3
	if c, ok := context["credit_utilization"].(float64); ok {
		creditUtilization = c
	}
	savingsRatio := 0.2
	if s, ok := context["savings_ratio"].(float64); ok {
		savingsRatio = s
	}

	payload := map[string]interface{}{
		"account_age_days":        accountAge,
		"monthly_income":          income,
		"total_balance":          balance,
		"transaction_count_30d":   txnCount30d,
		"delinquency_count":      delinquency,
		"loan_history_count":     loanHistory,
		"avg_transaction_amount": avgTxnAmount,
		"credit_utilization":     creditUtilization,
		"savings_ratio":          savingsRatio,
	}

	result, err := sa.CallMLService(ctx, "/api/v1/scoring/credit", payload)
	if err != nil {
		return nil, 0, "", err
	}

	// Extract credit score from response
	if resultData, ok := result["result"].(map[string]interface{}); ok {
		creditScore := 600.0
		if cs, ok := resultData["credit_score"].(float64); ok {
			creditScore = cs
		} else if cs, ok := resultData["credit_score"].(int); ok {
			creditScore = float64(cs)
		}

		riskScore := 1.0 - (creditScore / 850.0)
		scoreRange := sa.getScoreRange(creditScore)
		riskCategory := sa.getRiskCategory(riskScore)

		return map[string]interface{}{
			"credit_score":  int(creditScore),
			"score_range":   scoreRange,
			"risk_category": riskCategory,
			"factors": map[string]interface{}{
				"account_age":    accountAge,
				"income":         income,
				"delinquency":    delinquency,
				"loan_history":   loanHistory,
				"balance":        balance,
			},
		}, riskScore, fmt.Sprintf("Credit score calculated: %d (%s)", int(creditScore), scoreRange), nil
	}

	return nil, 0, "", fmt.Errorf("invalid response format from ML service")
}

// calculateCreditScoreFallback uses rule-based calculation as fallback
func (sa *ScoringAgent) calculateCreditScoreFallback(context map[string]interface{}) (map[string]interface{}, float64, string) {
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
	// Try to call ML Models service first
	if sa.mlModelsEnabled {
		result, riskScore, explanation, err := sa.callMLRiskModel(ctx, context)
		if err == nil {
			log.Info().Msg("Risk score from ML model")
			return result, riskScore, explanation
		}
		log.Warn().Err(err).Msg("ML model call failed, using fallback")
	}

	// Fallback to rule-based calculation
	return sa.calculateRiskScoreFallback(context)
}

// callMLRiskModel calls the ML Models service for risk scoring
func (sa *ScoringAgent) callMLRiskModel(ctx context.Context, context map[string]interface{}) (map[string]interface{}, float64, string, error) {
	// Extract features for risk model (combines credit + fraud features)
	payload := map[string]interface{}{
		"account_age_days":        context["account_age_days"],
		"monthly_income":          context["income"],
		"total_balance":          context["balance"],
		"transaction_count_30d":   context["transaction_count_30d"],
		"delinquency_count":      context["delinquency_count"],
		"loan_history_count":     context["loan_history_count"],
		"amount":                 context["amount"],
		"hour":                   context["hour"],
		"day_of_week":            context["day_of_week"],
		"transaction_count_24h":  context["transaction_count_24h"],
		"transaction_count_7d":    context["transaction_count_7d"],
		"avg_amount_7d":           context["avg_amount_7d"],
		"beneficiary_age_days":    context["beneficiary_age_days"],
		"device_risk":            context["device_risk"],
		"location_risk":          context["location_risk"],
	}

	// Set defaults for missing values
	defaults := map[string]float64{
		"account_age_days":        365.0,
		"monthly_income":          50000.0,
		"total_balance":           100000.0,
		"transaction_count_30d":   10.0,
		"delinquency_count":       0.0,
		"loan_history_count":      0.0,
		"amount":                  0.0,
		"hour":                    12.0,
		"day_of_week":             3.0,
		"transaction_count_24h":   0.0,
		"transaction_count_7d":    5.0,
		"avg_amount_7d":           10000.0,
		"beneficiary_age_days":    365.0,
		"device_risk":             0.0,
		"location_risk":           0.0,
	}

	for key, defaultValue := range defaults {
		if payload[key] == nil {
			payload[key] = defaultValue
		} else if _, ok := payload[key].(float64); !ok {
			payload[key] = defaultValue
		}
	}

	result, err := sa.CallMLService(ctx, "/api/v1/scoring/risk", payload)
	if err != nil {
		return nil, 0, "", err
	}

	// Extract risk score from response
	if resultData, ok := result["result"].(map[string]interface{}); ok {
		overallRisk := 0.5
		if or, ok := resultData["overall_risk"].(float64); ok {
			overallRisk = or
		}

		riskCategory := sa.getRiskCategory(overallRisk)

		return map[string]interface{}{
			"overall_risk":  overallRisk,
			"risk_category": riskCategory,
			"components":    resultData["components"],
		}, overallRisk, fmt.Sprintf("Overall risk score: %.2f (%s)", overallRisk, riskCategory), nil
	}

	return nil, 0, "", fmt.Errorf("invalid response format from ML service")
}

// calculateRiskScoreFallback uses rule-based calculation as fallback
func (sa *ScoringAgent) calculateRiskScoreFallback(context map[string]interface{}) (map[string]interface{}, float64, string) {
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

