package service

import (
	"context"

	"github.com/aibanking/ai-skin-orchestrator/internal/model"
)

// BehaviorAnalyzer analyzes user behavior patterns
type BehaviorAnalyzer struct {
}

// NewBehaviorAnalyzer creates a new behavior analyzer
func NewBehaviorAnalyzer() *BehaviorAnalyzer {
	return &BehaviorAnalyzer{}
}

// AnalyzeBehavior analyzes transaction history to identify behavior patterns
func (ba *BehaviorAnalyzer) AnalyzeBehavior(ctx context.Context, userID string, history []model.TransactionRecord) model.BehaviorPattern {
	if len(history) == 0 {
		return model.BehaviorPattern{
			AverageAmount:        0,
			PeakHours:            []int{},
			CommonChannels:        []string{},
			FrequentBeneficiaries: []string{},
			AnomalyDetected:      false,
		}
	}

	// Calculate average amount
	var totalAmount float64
	for _, txn := range history {
		totalAmount += txn.Amount
	}
	averageAmount := totalAmount / float64(len(history))

	// Analyze peak hours (mock - in production would analyze timestamps)
	peakHours := []int{10, 11, 14, 15, 20, 21}

	// Common channels (mock)
	commonChannels := []string{"MB", "NB"}

	// Frequent beneficiaries (mock)
	frequentBeneficiaries := []string{"XXXX4321", "YYYY5678"}

	// Detect anomalies (simple check - amount > 2x average)
	anomalyDetected := false
	for _, txn := range history {
		if txn.Amount > averageAmount*2 {
			anomalyDetected = true
			break
		}
	}

	return model.BehaviorPattern{
		AverageAmount:         averageAmount,
		PeakHours:            peakHours,
		CommonChannels:        commonChannels,
		FrequentBeneficiaries: frequentBeneficiaries,
		AnomalyDetected:      anomalyDetected,
	}
}

