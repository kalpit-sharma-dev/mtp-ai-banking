# ML Models Integration - Current Issue & Fix

## 🚨 Current Problem

**ML Models are NOT being called!** The agents are using rule-based calculations instead of calling the actual ML Models service.

### Current Implementation (WRONG)

#### Fraud Agent (`fraud_agent.go`)
```go
// calculateFraudScore calculates fraud risk score using ML model simulation
func (fa *FraudAgent) calculateFraudScore(...) float64 {
    score := 0.0
    
    // Amount-based risk (RULE-BASED, NOT ML)
    if amount > 200000 {
        score += 0.4
    } else if amount > 100000 {
        score += 0.2
    }
    // ... more rule-based logic
    
    return score  // ❌ NOT calling ML Models service
}
```

#### Scoring Agent (`scoring_agent.go`)
```go
// calculateCreditScore calculates credit score
func (sa *ScoringAgent) calculateCreditScore(...) {
    score := 600.0  // ❌ Rule-based calculation
    
    // Account age factor (RULE-BASED, NOT ML)
    if accountAge > 365 {
        score += 50
    }
    // ... more rule-based logic
    
    return result, riskScore, explanation  // ❌ NOT calling ML Models service
}
```

### What Should Happen

The agents should call the **ML Models Service (Port 9000)**:

```
Fraud Agent → POST http://localhost:9000/api/v1/fraud/predict
Scoring Agent → POST http://localhost:9000/api/v1/scoring/credit
Scoring Agent → POST http://localhost:9000/api/v1/scoring/risk
```

---

## ✅ Solution: Integrate ML Models Service

### Step 1: Add HTTP Client to Agents

Agents need to make HTTP calls to ML Models service.

### Step 2: Update Fraud Agent

Replace rule-based calculation with ML model call:

```go
func (fa *FraudAgent) calculateFraudScore(ctx context.Context, amount float64, toAccount string, userID string, context map[string]interface{}) float64 {
    // Prepare features for ML model
    features := map[string]interface{}{
        "amount": amount,
        "hour": time.Now().Hour(),
        "day_of_week": int(time.Now().Weekday()),
        "transaction_count_24h": context["transaction_count_24h"],
        "transaction_count_7d": context["transaction_count_7d"],
        "beneficiary_age_days": context["beneficiary_age_days"],
        "device_risk": context["device_risk"],
        "location_risk": context["location_risk"],
        "user_account_age_days": context["user_account_age_days"],
        "user_balance": context["user_balance"],
    }
    
    // Call ML Models Service
    mlResponse, err := fa.callMLFraudModel(ctx, features)
    if err != nil {
        log.Warn().Err(err).Msg("ML model call failed, using fallback")
        return fa.calculateFraudScoreFallback(amount, toAccount, userID, context)
    }
    
    return mlResponse.FraudScore
}

func (fa *FraudAgent) callMLFraudModel(ctx context.Context, features map[string]interface{}) (*FraudPredictionResponse, error) {
    url := "http://localhost:9000/api/v1/fraud/predict"
    
    req := FraudPredictionRequest{
        Amount: features["amount"].(float64),
        Hour: features["hour"].(int),
        // ... map all features
    }
    
    resp, err := fa.httpClient.Post(url, req)
    // Parse response and return fraud score
}
```

### Step 3: Update Scoring Agent

Replace rule-based calculation with ML model call:

```go
func (sa *ScoringAgent) calculateCreditScore(ctx context.Context, context map[string]interface{}) (map[string]interface{}, float64, string) {
    // Prepare features for ML model
    features := CreditScoringRequest{
        AccountAgeDays: context["account_age_days"].(int),
        MonthlyIncome: context["income"].(float64),
        TotalBalance: context["balance"].(float64),
        // ... map all features
    }
    
    // Call ML Models Service
    mlResponse, err := sa.callMLCreditModel(ctx, features)
    if err != nil {
        log.Warn().Err(err).Msg("ML model call failed, using fallback")
        return sa.calculateCreditScoreFallback(context)
    }
    
    return mlResponse, riskScore, explanation
}

func (sa *ScoringAgent) callMLCreditModel(ctx context.Context, features CreditScoringRequest) (*CreditScoringResponse, error) {
    url := "http://localhost:9000/api/v1/scoring/credit"
    
    resp, err := sa.httpClient.Post(url, features)
    // Parse response and return credit score
}
```

---

## 📊 Updated Flow with ML Models

```
User: "Transfer 50000 rupees"
    ↓
AI Orchestrator → Intent: TRANSFER_NEFT
    ↓
MCP Server → Route to Fraud Agent
    ↓
Fraud Agent (Port 8002)
    ├─ Extract features
    └─ POST http://localhost:9000/api/v1/fraud/predict  ✅ ML MODEL CALL
        ↓
    ML Models Service (Port 9000)
        ├─ Load XGBoost model
        ├─ Predict fraud probability
        └─ Return: {fraud_score: 0.15, risk_level: "LOW"}
        ↓
    Fraud Agent
        └─ Return fraud score to MCP Server
            ↓
    MCP Server → Route to Banking Agent
        ↓
    Banking Agent → Process transfer
```

---

## 🔧 Implementation Steps

1. **Add HTTP client to AgentBase**
2. **Add ML Models service URL to config**
3. **Update Fraud Agent to call ML service**
4. **Update Scoring Agent to call ML service**
5. **Add fallback logic (if ML service unavailable)**
6. **Update flow documentation**

---

## Current vs. Correct Flow

### ❌ Current (Wrong)
```
Fraud Agent → Rule-based calculation → Return score
```

### ✅ Correct (Should Be)
```
Fraud Agent → HTTP POST to ML Models → XGBoost prediction → Return score
```

---

## Why This Matters

- **ML Models are trained** on real data patterns
- **Rule-based** is just simple if-else logic
- **ML Models** provide better accuracy and fraud detection
- **ML Models** can learn from new patterns

The ML Models service exists and is ready, but agents are not calling it!

