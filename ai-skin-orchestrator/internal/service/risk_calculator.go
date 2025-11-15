package service

import (
	"context"
	"fmt"

	"github.com/aibanking/ai-skin-orchestrator/internal/model"
)

// RiskCalculator calculates risk indicators
type RiskCalculator struct {
}

// NewRiskCalculator creates a new risk calculator
func NewRiskCalculator() *RiskCalculator {
	return &RiskCalculator{}
}

// CalculateRisk calculates risk indicators based on intent, history, and behavior
func (rc *RiskCalculator) CalculateRisk(
	ctx context.Context,
	userID string,
	intent model.Intent,
	history []model.TransactionRecord,
	behavior model.BehaviorPattern,
) model.RiskIndicators {
	// Extract amount from intent entities
	var amount float64
	if amountVal, ok := intent.Entities["amount"]; ok {
		if amountStr, ok := amountVal.(string); ok {
			// Parse amount string
			amount = parseAmount(amountStr)
		} else if amountFloat, ok := amountVal.(float64); ok {
			amount = amountFloat
		}
	}

	// Calculate fraud risk
	fraudRisk := rc.calculateFraudRisk(amount, behavior)

	// Calculate credit risk
	creditRisk := rc.calculateCreditRisk(amount, behavior)

	// Calculate velocity risk (too many transactions)
	velocityRisk := rc.calculateVelocityRisk(len(history))

	// Calculate amount risk
	amountRisk := rc.calculateAmountRisk(amount, behavior.AverageAmount)

	// Overall risk
	overallRisk := "LOW"
	if fraudRisk > 0.7 || creditRisk > 0.7 || amountRisk > 0.7 {
		overallRisk = "HIGH"
	} else if fraudRisk > 0.4 || creditRisk > 0.4 || amountRisk > 0.4 {
		overallRisk = "MEDIUM"
	}

	return model.RiskIndicators{
		OverallRisk:  overallRisk,
		FraudRisk:    fraudRisk,
		CreditRisk:   creditRisk,
		VelocityRisk: velocityRisk,
		AmountRisk:   amountRisk,
		DeviceRisk:   0.1, // Mock
		LocationRisk: 0.1, // Mock
	}
}

func (rc *RiskCalculator) calculateFraudRisk(amount float64, behavior model.BehaviorPattern) float64 {
	risk := 0.1

	// High amount increases risk
	if amount > 100000 {
		risk += 0.4
	} else if amount > 50000 {
		risk += 0.2
	}

	// Anomaly detection
	if behavior.AnomalyDetected {
		risk += 0.3
	}

	// Amount significantly different from average
	if amount > 0 && behavior.AverageAmount > 0 {
		if amount > behavior.AverageAmount*3 {
			risk += 0.2
		}
	}

	if risk > 1.0 {
		risk = 1.0
	}

	return risk
}

func (rc *RiskCalculator) calculateCreditRisk(amount float64, behavior model.BehaviorPattern) float64 {
	risk := 0.1

	// Very high amounts
	if amount > 200000 {
		risk += 0.3
	}

	return risk
}

func (rc *RiskCalculator) calculateVelocityRisk(transactionCount int) float64 {
	if transactionCount > 50 {
		return 0.7
	} else if transactionCount > 30 {
		return 0.4
	}
	return 0.1
}

func (rc *RiskCalculator) calculateAmountRisk(amount, averageAmount float64) float64 {
	if amount == 0 || averageAmount == 0 {
		return 0.1
	}

	ratio := amount / averageAmount
	if ratio > 5 {
		return 0.8
	} else if ratio > 3 {
		return 0.5
	} else if ratio > 2 {
		return 0.3
	}
	return 0.1
}

func parseAmount(amountStr string) float64 {
	// Simple parser - in production would handle currency symbols, commas, etc.
	var amount float64
	_, err := fmt.Sscanf(amountStr, "%f", &amount)
	if err != nil {
		return 0
	}
	return amount
}

