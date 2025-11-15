# AI & ML Models Used in the Banking Platform

## Overview

The AI Banking Platform uses **two types of AI/ML models**:

1. **Machine Learning Models** (Layer 4) - For fraud detection and scoring
2. **Large Language Models (LLMs)** (Layer 2) - For natural language understanding

---

## 1. Machine Learning Models (Layer 4: ML Models Service)

### Location: `ml-models/` directory

### A. Fraud Detection Model
- **Algorithm**: **XGBoost Classifier** (Gradient Boosting)
- **Library**: `xgboost==2.0.3`
- **Purpose**: Predicts fraud probability for banking transactions
- **Output**: 
  - Fraud score (0.0 to 1.0)
  - Risk level (LOW, MEDIUM, HIGH, CRITICAL)
  - Fraud flag (boolean)
- **Features Used**:
  - Transaction amount
  - Time features (hour, day of week)
  - Transaction velocity (24h, 7d counts)
  - Beneficiary age (days since added)
  - Device and location risk
  - User account information
- **File**: `ml-models/app/models/fraud_model.py`
- **Training**: `ml-models/train_models.py` (generates synthetic data and trains model)
- **Model File**: `models/fraud_detection_model.pkl` (saved after training)

### B. Credit Scoring Model
- **Algorithm**: **Random Forest Regressor** (Ensemble Learning)
- **Library**: `scikit-learn==1.3.2`
- **Purpose**: Predicts credit score (300-850 range)
- **Output**:
  - Credit score (300-850)
  - Risk category (EXCELLENT, GOOD, FAIR, POOR)
  - Score range
  - Factor analysis
- **Features Used**:
  - Account age (months)
  - Monthly income
  - Total balance
  - Transaction history (count, average amount)
  - Delinquency count
  - Loan history
  - Credit utilization
  - Savings ratio
- **File**: `ml-models/app/models/credit_model.py`
- **Training**: `ml-models/train_models.py`
- **Model File**: `models/credit_scoring_model.pkl`

### C. Risk Scoring Model
- **Algorithm**: **Ensemble Model** (Weighted combination)
- **Purpose**: Overall risk assessment combining multiple factors
- **Method**: 
  - Combines Credit Risk (40%) + Fraud Risk (40%) + Amount Risk (20%)
  - Uses outputs from both Fraud and Credit models
- **Output**:
  - Overall risk score (0.0 to 1.0)
  - Risk category (LOW, MEDIUM, HIGH, CRITICAL)
  - Recommendation (APPROVE, REVIEW, REJECT)
- **File**: `ml-models/app/models/risk_model.py`
- **No Training Required**: Always available, combines other models

### ML Libraries Used:
```python
scikit-learn==1.3.2    # Random Forest, model utilities
xgboost==2.0.3         # Gradient Boosting for fraud detection
numpy==1.24.3          # Numerical computations
pandas==2.1.3          # Data manipulation
joblib==1.3.2          # Model serialization
```

### Model Training:
- **Script**: `ml-models/train_models.py`
- **Data**: Synthetic data generation (10,000 samples)
- **Fallback**: Models work without training using rule-based mock predictions

---

## 2. Large Language Models (LLM) (Layer 2: AI Skin Orchestrator)

### Location: `ai-skin-orchestrator/` directory

### LLM Service
- **Provider**: **OpenAI** (default)
- **Default Model**: **GPT-3.5-turbo**
- **Library**: `github.com/sashabaranov/go-openai v1.20.0`
- **Purpose**: Natural language understanding and intent parsing
- **Usage**:
  - Parses user natural language requests
  - Extracts banking intents (TRANSFER, CHECK_BALANCE, etc.)
  - Extracts entities (amount, account numbers, etc.)
  - Provides confidence scores

### Configuration:
- **Model**: Configurable (default: `gpt-3.5-turbo`)
- **Temperature**: 0.7 (default)
- **Max Tokens**: 1000 (default)
- **Base URL**: Supports custom/self-hosted models (e.g., local LLM servers)
- **Optional**: Falls back to rule-based parsing if LLM is disabled

### File: `ai-skin-orchestrator/internal/service/llm_service.go`

### Supported Models:
- **OpenAI**: GPT-3.5-turbo, GPT-4, GPT-4-turbo (via API)
- **Custom/Self-hosted**: Any OpenAI-compatible API (via BaseURL)
  - Examples: Local Llama, Mistral, or other OpenAI-compatible servers

### Environment Variables:
```bash
LLM_ENABLED=true                    # Enable/disable LLM
LLM_PROVIDER=openai                 # Provider (openai, anthropic, local)
LLM_API_KEY=your-api-key            # API key for OpenAI
LLM_MODEL=gpt-3.5-turbo            # Model name
LLM_TEMPERATURE=0.7                 # Temperature (0.0-2.0)
LLM_MAX_TOKENS=1000                 # Max tokens
LLM_BASE_URL=                       # Custom base URL for self-hosted models
```

### Fallback Behavior:
- If LLM is disabled or unavailable → Uses rule-based intent parsing
- Rule-based parser uses keyword matching and regex patterns
- File: `ai-skin-orchestrator/internal/service/intent_parser.go`

---

## Model Architecture Flow

```
User Request (Natural Language)
        |
        v
AI Skin Orchestrator (Layer 2)
  ├── LLM Service (GPT-3.5-turbo) → Intent Parsing
  └── Rule-based Parser (Fallback)
        |
        v
MCP Server (Layer 1)
        |
        v
Agent Mesh (Layer 3)
  ├── Fraud Agent → ML Models Service (Layer 4)
  │                    └── Fraud Detection (XGBoost)
  ├── Scoring Agent → ML Models Service (Layer 4)
  │                    ├── Credit Scoring (Random Forest)
  │                    └── Risk Scoring (Ensemble)
  └── Other Agents
```

---

## Summary

### ML Models (Traditional Machine Learning):
1. ✅ **XGBoost Classifier** - Fraud Detection
2. ✅ **Random Forest Regressor** - Credit Scoring
3. ✅ **Ensemble Model** - Risk Scoring

### AI Models (Large Language Models):
1. ✅ **OpenAI GPT-3.5-turbo** (default) - Natural Language Understanding
2. ✅ **Custom/Self-hosted LLMs** (via OpenAI-compatible API)

### Key Features:
- **ML Models**: Trained on synthetic data, can be retrained with real data
- **LLM**: Optional, falls back to rule-based if disabled
- **Mock Fallback**: All models work without training files using rule-based predictions
- **Production Ready**: All models are production-ready and can be deployed

---

## How to Use

### Train ML Models:
```bash
cd ml-models
python train_models.py
```

### Enable LLM:
```bash
# In ai-skin-orchestrator/.env
LLM_ENABLED=true
LLM_API_KEY=your-openai-api-key
LLM_MODEL=gpt-3.5-turbo
```

### Use Self-hosted LLM:
```bash
# In ai-skin-orchestrator/.env
LLM_ENABLED=true
LLM_BASE_URL=http://localhost:8000/v1  # Your local LLM server
LLM_API_KEY=dummy-key
LLM_MODEL=llama-2-7b
```

---

## Notes

- **ML Models** are trained on synthetic data by default. For production, retrain with real banking data.
- **LLM** is optional. The system works perfectly fine with rule-based parsing.
- **All models** have mock fallbacks, so the system works even without trained models or LLM API keys.
- **Model files** are saved in `ml-models/models/` directory after training.

