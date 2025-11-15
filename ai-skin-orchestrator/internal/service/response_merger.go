package service

import (
	"fmt"

	"github.com/aibanking/ai-skin-orchestrator/internal/model"
)

// ResponseMerger merges responses from multiple agents
type ResponseMerger struct {
}

// NewResponseMerger creates a new response merger
func NewResponseMerger() *ResponseMerger {
	return &ResponseMerger{}
}

// MergeResponses merges multiple agent responses into a single response
func (rm *ResponseMerger) MergeResponses(responses []model.AgentResponse) (*model.MergedResponse, error) {
	if len(responses) == 0 {
		return nil, fmt.Errorf("no responses to merge")
	}

	if len(responses) == 1 {
		return rm.singleResponseToMerged(responses[0]), nil
	}

	// Check for conflicts
	conflicts := rm.detectConflicts(responses)

	// Determine final status
	finalStatus := rm.determineFinalStatus(responses, conflicts)

	// Merge results
	finalResult := rm.mergeResults(responses)

	// Calculate average risk score
	avgRiskScore := rm.calculateAverageRiskScore(responses)

	// Generate explanation
	explanation := rm.generateExplanation(responses, conflicts)

	// Determine who resolved conflicts
	resolvedBy := ""
	if len(conflicts) > 0 {
		resolvedBy = rm.resolveConflicts(responses, conflicts)
	}

	return &model.MergedResponse{
		Status:         finalStatus,
		FinalResult:    finalResult,
		RiskScore:      avgRiskScore,
		Explanation:    explanation,
		AgentResponses: responses,
		Conflicts:      conflicts,
		ResolvedBy:     resolvedBy,
	}, nil
}

// singleResponseToMerged converts a single response to merged format
func (rm *ResponseMerger) singleResponseToMerged(resp model.AgentResponse) *model.MergedResponse {
	return &model.MergedResponse{
		Status:         resp.Status,
		FinalResult:    resp.Result,
		RiskScore:      resp.RiskScore,
		Explanation:    resp.Explanation,
		AgentResponses: []model.AgentResponse{resp},
	}
}

// detectConflicts detects conflicts between agent responses
func (rm *ResponseMerger) detectConflicts(responses []model.AgentResponse) []model.Conflict {
	var conflicts []model.Conflict

	// Check for status mismatches
	statuses := make(map[string][]string)
	for _, resp := range responses {
		statuses[resp.Status] = append(statuses[resp.Status], resp.AgentID)
	}

	if len(statuses) > 1 {
		conflict := model.Conflict{
			Type:        "STATUS_MISMATCH",
			Description: "Agents returned different statuses",
			Agents:      []string{},
			Values:      make(map[string]interface{}),
		}
		for status, agents := range statuses {
			conflict.Agents = append(conflict.Agents, agents...)
			conflict.Values[status] = agents
		}
		conflicts = append(conflicts, conflict)
	}

	// Check for significant risk score differences
	if len(responses) > 1 {
		var riskScores []float64
		for _, resp := range responses {
			riskScores = append(riskScores, resp.RiskScore)
		}

		maxRisk := riskScores[0]
		minRisk := riskScores[0]
		for _, risk := range riskScores {
			if risk > maxRisk {
				maxRisk = risk
			}
			if risk < minRisk {
				minRisk = risk
			}
		}

		if maxRisk-minRisk > 0.3 {
			conflict := model.Conflict{
				Type:        "RISK_SCORE_MISMATCH",
				Description: fmt.Sprintf("Risk scores vary significantly (%.2f to %.2f)", minRisk, maxRisk),
				Agents:      []string{},
				Values: map[string]interface{}{
					"min_risk": minRisk,
					"max_risk": maxRisk,
				},
			}
			for _, resp := range responses {
				conflict.Agents = append(conflict.Agents, resp.AgentID)
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts
}

// determineFinalStatus determines the final status based on responses and conflicts
func (rm *ResponseMerger) determineFinalStatus(responses []model.AgentResponse, conflicts []model.Conflict) string {
	if len(conflicts) > 0 {
		// If there are conflicts, check if any agent rejected
		for _, resp := range responses {
			if resp.Status == "REJECTED" {
				return "REJECTED"
			}
		}
		return "CONFLICT"
	}

	// If no conflicts, use majority vote or most restrictive
	rejectedCount := 0
	approvedCount := 0
	for _, resp := range responses {
		if resp.Status == "REJECTED" {
			rejectedCount++
		} else if resp.Status == "APPROVED" {
			approvedCount++
		}
	}

	if rejectedCount > 0 {
		return "REJECTED"
	}
	if approvedCount > 0 {
		return "APPROVED"
	}

	return "PENDING"
}

// mergeResults merges result maps from all agents
func (rm *ResponseMerger) mergeResults(responses []model.AgentResponse) map[string]interface{} {
	merged := make(map[string]interface{})

	for _, resp := range responses {
		for k, v := range resp.Result {
			// If key already exists, prefer the one with higher confidence
			if existing, exists := merged[k]; exists {
				if resp.Confidence > 0.8 {
					merged[k] = v
				} else {
					merged[k] = existing
				}
			} else {
				merged[k] = v
			}
		}
	}

	return merged
}

// calculateAverageRiskScore calculates average risk score
func (rm *ResponseMerger) calculateAverageRiskScore(responses []model.AgentResponse) float64 {
	if len(responses) == 0 {
		return 0.0
	}

	var sum float64
	for _, resp := range responses {
		sum += resp.RiskScore
	}

	return sum / float64(len(responses))
}

// generateExplanation generates a merged explanation
func (rm *ResponseMerger) generateExplanation(responses []model.AgentResponse, conflicts []model.Conflict) string {
	if len(conflicts) > 0 {
		return fmt.Sprintf("Multiple agents evaluated the request. Conflicts detected and resolved. %d agent(s) participated.", len(responses))
	}

	if len(responses) == 1 {
		return responses[0].Explanation
	}

	return fmt.Sprintf("Request evaluated by %d agents. All agents agree on the decision.", len(responses))
}

// resolveConflicts resolves conflicts using a strategy (e.g., highest confidence, most restrictive)
func (rm *ResponseMerger) resolveConflicts(responses []model.AgentResponse, conflicts []model.Conflict) string {
	// Strategy: Use the agent with highest confidence, or if status conflict, prefer REJECTED
	for _, resp := range responses {
		if resp.Status == "REJECTED" && resp.Confidence > 0.7 {
			return fmt.Sprintf("Agent %s (most restrictive)", resp.AgentID)
		}
	}

	// Otherwise, use highest confidence
	highestConf := 0.0
	bestAgent := ""
	for _, resp := range responses {
		if resp.Confidence > highestConf {
			highestConf = resp.Confidence
			bestAgent = resp.AgentID
		}
	}

	return fmt.Sprintf("Agent %s (highest confidence: %.2f)", bestAgent, highestConf)
}

