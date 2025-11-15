package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aibanking/agent-mesh/internal/model"
	"github.com/rs/zerolog/log"
)

// ClearanceAgent handles loan approval and clearance decisions
type ClearanceAgent struct {
	*AgentBase
}

// NewClearanceAgent creates a new clearance agent
func NewClearanceAgent(base *AgentBase) *ClearanceAgent {
	return &ClearanceAgent{
		AgentBase: base,
	}
}

// Process processes a clearance request
func (ca *ClearanceAgent) Process(ctx context.Context, req *model.AgentRequest) (*model.AgentResponse, error) {
	log.Info().
		Str("task", req.Task).
		Str("request_id", req.RequestID).
		Msg("Clearance agent processing request")

	inputCtx := req.InputContext
	data, ok := inputCtx["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data in input context")
	}

	loanType, _ := data["loan_type"].(string)
	amount, _ := data["amount"].(float64)
	tenure, _ := data["tenure"].(float64)
	_ = inputCtx["user_id"] // userID available in context if needed

	// Get user profile for clearance decision
	creditScore, _ := inputCtx["credit_score"].(float64)
	income, _ := inputCtx["income"].(float64)

	// Perform clearance checks
	clearanceDecision := ca.makeClearanceDecision(ctx, loanType, amount, tenure, creditScore, income, inputCtx)

	log.Info().
		Str("loan_type", loanType).
		Float64("amount", amount).
		Str("decision", clearanceDecision.Status).
		Msg("Clearance decision made")

	result := map[string]interface{}{
		"clearance_level": clearanceDecision.ClearanceLevel,
		"loan_amount":     clearanceDecision.ApprovedAmount,
		"interest_rate":   clearanceDecision.InterestRate,
		"tenure":          tenure,
		"conditions":     clearanceDecision.Conditions,
		"reason":         clearanceDecision.Reason,
	}

	return &model.AgentResponse{
		AgentID:     ca.agentType,
		AgentType:   "CLEARANCE",
		Status:      clearanceDecision.Status,
		Result:      result,
		RiskScore:   clearanceDecision.RiskScore,
		Explanation: clearanceDecision.Explanation,
		Confidence:  0.9,
		Timestamp:   time.Now(),
		RequestID:   req.RequestID,
	}, nil
}

// ClearanceDecision represents a clearance decision
type ClearanceDecision struct {
	Status        string
	ClearanceLevel string // AUTO, MANUAL, REJECTED
	ApprovedAmount float64
	InterestRate  float64
	RiskScore     float64
	Conditions    []string
	Reason        string
	Explanation   string
}

// makeClearanceDecision makes a clearance decision based on loan parameters
func (ca *ClearanceAgent) makeClearanceDecision(ctx context.Context, loanType string, amount, tenure, creditScore, income float64, context map[string]interface{}) ClearanceDecision {
	decision := ClearanceDecision{
		Status:         "APPROVED",
		ClearanceLevel: "AUTO",
		ApprovedAmount: amount,
		InterestRate:   8.5,
		RiskScore:      0.2,
		Conditions:     []string{},
		Reason:         "Loan approved based on credit score and income",
		Explanation:    "Loan application meets all clearance criteria",
	}

	// Credit score check
	if creditScore < 600 {
		decision.Status = "REJECTED"
		decision.ClearanceLevel = "REJECTED"
		decision.Reason = "Credit score below minimum threshold"
		decision.Explanation = "Credit score is too low for loan approval"
		decision.RiskScore = 1.0
		return decision
	} else if creditScore < 700 {
		decision.ClearanceLevel = "MANUAL"
		decision.Conditions = append(decision.Conditions, "MANUAL_REVIEW_REQUIRED")
		decision.RiskScore = 0.5
		decision.Reason = "Credit score requires manual review"
	}

	// Income-to-loan ratio check
	if income > 0 {
		emi := ca.calculateEMI(amount, decision.InterestRate, tenure)
		emiToIncomeRatio := (emi / income) * 100

		if emiToIncomeRatio > 50 {
			decision.Status = "REJECTED"
			decision.ClearanceLevel = "REJECTED"
			decision.Reason = "EMI exceeds 50% of income"
			decision.Explanation = "Loan EMI is too high compared to income"
			decision.RiskScore = 0.9
			return decision
		} else if emiToIncomeRatio > 40 {
			decision.ClearanceLevel = "MANUAL"
			decision.Conditions = append(decision.Conditions, "HIGH_EMI_RATIO")
			decision.RiskScore = 0.6
		}
	}

	// Loan amount limits based on type
	maxAmount := ca.getMaxLoanAmount(loanType, creditScore, income)
	if amount > maxAmount {
		decision.ApprovedAmount = maxAmount
		decision.Conditions = append(decision.Conditions, "AMOUNT_ADJUSTED")
		decision.Reason = fmt.Sprintf("Loan amount adjusted to maximum eligible: %.2f", maxAmount)
	}

	// Adjust interest rate based on credit score
	if creditScore >= 750 {
		decision.InterestRate = 7.5
	} else if creditScore >= 700 {
		decision.InterestRate = 8.5
	} else {
		decision.InterestRate = 10.0
		decision.Conditions = append(decision.Conditions, "HIGHER_INTEREST_RATE")
	}

	return decision
}

// calculateEMI calculates Equated Monthly Installment
func (ca *ClearanceAgent) calculateEMI(principal, rate, tenure float64) float64 {
	monthlyRate := rate / 12 / 100
	emi := principal * monthlyRate * (1 + monthlyRate) * tenure / ((1 + monthlyRate) * tenure - 1)
	return emi
}

// getMaxLoanAmount returns maximum loan amount based on loan type and profile
func (ca *ClearanceAgent) getMaxLoanAmount(loanType string, creditScore, income float64) float64 {
	baseMultiplier := 10.0 // Base: 10x monthly income

	// Adjust based on credit score
	if creditScore >= 750 {
		baseMultiplier = 15.0
	} else if creditScore >= 700 {
		baseMultiplier = 12.0
	}

	// Adjust based on loan type
	switch loanType {
	case "HOME":
		baseMultiplier *= 2.0
	case "AUTO":
		baseMultiplier *= 1.5
	case "PERSONAL":
		baseMultiplier *= 1.0
	}

	maxAmount := income * baseMultiplier

	// Cap based on loan type
	switch loanType {
	case "HOME":
		if maxAmount > 5000000 {
			maxAmount = 5000000
		}
	case "AUTO":
		if maxAmount > 2000000 {
			maxAmount = 2000000
		}
	case "PERSONAL":
		if maxAmount > 1000000 {
			maxAmount = 1000000
		}
	}

	return maxAmount
}

