# Layer 4: ML Models Service

The ML Models Service provides machine learning models for fraud detection and credit/risk scoring. This service can be called by agents in the Agent Mesh (Layer 3) for ML-based predictions.

## Models

### 1. Fraud Detection Model
- **Algorithm**: XGBoost Classifier
- **Purpose**: Predicts fraud probability for transactions
- **Output**: Fraud score (0.0 to 1.0), risk level, fraud flag
- **Features**: Amount, transaction patterns, beneficiary info, device/location risk

### 2. Credit Scoring Model
- **Algorithm**: Random Forest Regressor
- **Purpose**: Predicts credit score (300-850 range)
- **Output**: Credit score, risk category, score range
- **Features**: Account age, income, balance, transaction history, delinquency

### 3. Risk Scoring Model
- **Algorithm**: Ensemble (combines credit + fraud)
- **Purpose**: Overall risk assessment
- **Output**: Overall risk score, category, recommendation
- **Features**: Combines credit and fraud features

## Architecture

```
Agent Mesh (Layer 3)
  ├── Fraud Agent
  ├── Scoring Agent
  └── Other Agents
        |
        v
ML Models Service (Layer 4)
  ├── Fraud Detection Model
  ├── Credit Scoring Model
  └── Risk Scoring Model
```

## Installation

1. Navigate to the ml-models directory:
```bash
cd ml-models
```

2. Create virtual environment (recommended):
```bash
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

3. Install dependencies:
```bash
pip install -r requirements.txt
```

4. Train models (optional - models work without training using mock predictions):
```bash
python train_models.py
```

5. Copy environment file:
```bash
cp .env.example .env
```

6. Run the service:
```bash
python -m app.main
# OR
uvicorn app.main:app --host 0.0.0.0 --port 9000
```

The service will start on `http://localhost:9000`

## API Endpoints

### Fraud Detection

**POST** `/api/v1/fraud/predict`

Predicts fraud probability for a transaction.

**Request:**
```json
{
  "amount": 50000,
  "hour": 14.0,
  "transaction_count_24h": 3.0,
  "beneficiary_age_days": 5.0,
  "device_risk": 0.2,
  "location_risk": 0.1
}
```

**Response:**
```json
{
  "success": true,
  "result": {
    "fraud_score": 0.35,
    "risk_level": "MEDIUM",
    "is_fraud": false
  }
}
```

### Credit Scoring

**POST** `/api/v1/scoring/credit`

Predicts credit score.

**Request:**
```json
{
  "account_age_days": 365,
  "monthly_income": 50000,
  "total_balance": 150000,
  "delinquency_count": 0,
  "loan_history_count": 2
}
```

**Response:**
```json
{
  "success": true,
  "result": {
    "credit_score": 750,
    "risk_category": "LOW",
    "score_range": "EXCELLENT",
    "factors": {
      "account_age": "Established account - positive impact",
      "income": "High income - positive impact"
    }
  }
}
```

### Risk Scoring

**POST** `/api/v1/scoring/risk`

Predicts overall risk score.

**Request:**
```json
{
  "account_age_days": 365,
  "monthly_income": 50000,
  "total_balance": 150000,
  "amount": 50000,
  "transaction_count_24h": 2.0,
  "beneficiary_age_days": 30.0
}
```

**Response:**
```json
{
  "success": true,
  "result": {
    "overall_risk_score": 0.25,
    "risk_category": "LOW",
    "components": {
      "credit_risk": 0.12,
      "fraud_risk": 0.15,
      "amount_risk": 0.3
    },
    "credit_score": 750,
    "fraud_score": 0.15,
    "recommendation": "APPROVE"
  }
}
```

### Health Check

**GET** `/health`

Returns service health status.

## Model Training

To train the models with synthetic data:

```bash
python train_models.py
```

This will:
1. Generate synthetic training data
2. Train fraud detection model (XGBoost)
3. Train credit scoring model (Random Forest)
4. Save models to `models/` directory

**Note**: Models work without training files using mock predictions. Training improves accuracy.

## Integration with Agents

Agents in Layer 3 can call this service:

```python
import requests

# Fraud prediction
response = requests.post(
    "http://localhost:9000/api/v1/fraud/predict",
    json={
        "amount": 50000,
        "transaction_count_24h": 3.0,
        "beneficiary_age_days": 5.0
    }
)
fraud_score = response.json()["result"]["fraud_score"]
```

## Configuration

### Environment Variables

- **HOST**: Server host (default: 0.0.0.0)
- **PORT**: Server port (default: 9000)
- **FRAUD_MODEL_PATH**: Path to fraud model file
- **CREDIT_MODEL_PATH**: Path to credit model file
- **FRAUD_THRESHOLD**: Fraud detection threshold (default: 0.5)

## Model Features

### Fraud Detection Features
- Transaction amount
- Time features (hour, day of week)
- Transaction velocity (24h, 7d counts)
- Beneficiary information
- Device and location risk
- User account information

### Credit Scoring Features
- Account age
- Monthly income
- Total balance
- Transaction history
- Delinquency count
- Loan history
- Credit utilization
- Savings ratio

## Production Considerations

1. **Model Versioning**: Implement model versioning for updates
2. **A/B Testing**: Test new models before full deployment
3. **Monitoring**: Monitor model performance and drift
4. **Retraining**: Schedule periodic retraining with new data
5. **Feature Store**: Use a feature store for consistent feature engineering
6. **Model Registry**: Store models in a model registry (MLflow, etc.)

## Next Steps

This is **Layer 4: ML Models**. The next layers to build:

- **Layer 5**: Banking Integrations (MB, NB, DWH connections)

## License

[Your License Here]

