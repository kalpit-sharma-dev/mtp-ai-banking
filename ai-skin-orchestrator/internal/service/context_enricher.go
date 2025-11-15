package service

import (
	"context"
	"time"

	"github.com/aibanking/ai-skin-orchestrator/internal/model"
	"github.com/rs/zerolog/log"
)

// ContextEnricher enriches context with user history and behavior patterns
type ContextEnricher struct {
	historyService *HistoryService
	behaviorAnalyzer *BehaviorAnalyzer
	riskCalculator *RiskCalculator
}

// NewContextEnricher creates a new context enricher
func NewContextEnricher(
	historyService *HistoryService,
	behaviorAnalyzer *BehaviorAnalyzer,
	riskCalculator *RiskCalculator,
) *ContextEnricher {
	return &ContextEnricher{
		historyService:   historyService,
		behaviorAnalyzer: behaviorAnalyzer,
		riskCalculator:   riskCalculator,
	}
}

// EnrichContext enriches context with user profile, history, and patterns
func (ce *ContextEnricher) EnrichContext(ctx context.Context, userID, sessionID, channel string, intent model.Intent) (*model.EnrichedContext, error) {
	// Get user profile (mock for now, in production would query database)
	userProfile := ce.getUserProfile(ctx, userID)

	// Get transaction history
	history, err := ce.historyService.GetTransactionHistory(ctx, userID, 90)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get transaction history")
		history = []model.TransactionRecord{}
	}

	// Analyze behavior patterns
	behaviorPattern := ce.behaviorAnalyzer.AnalyzeBehavior(ctx, userID, history)

	// Calculate risk indicators
	riskIndicators := ce.riskCalculator.CalculateRisk(ctx, userID, intent, history, behaviorPattern)

	enriched := &model.EnrichedContext{
		UserID:            userID,
		SessionID:         sessionID,
		Channel:           channel,
		Intent:            intent,
		UserProfile:       userProfile,
		TransactionHistory: history,
		RiskIndicators:    riskIndicators,
		BehaviorPattern:  behaviorPattern,
		Metadata:          make(map[string]interface{}),
	}

	// Add metadata
	enriched.Metadata["enrichment_timestamp"] = time.Now()
	enriched.Metadata["history_count"] = len(history)
	enriched.Metadata["channel"] = channel

	return enriched, nil
}

// getUserProfile retrieves user profile (mock implementation)
func (ce *ContextEnricher) getUserProfile(ctx context.Context, userID string) model.UserProfile {
	// In production, this would query a user database
	// For now, return mock data
	return model.UserProfile{
		AccountAge:       365,
		TotalBalance:     150000.0,
		MonthlyIncome:    50000.0,
		CreditScore:      750,
		KYCStatus:         "VERIFIED",
		AccountType:       "SAVINGS",
		TransactionCount: 25,
	}
}

