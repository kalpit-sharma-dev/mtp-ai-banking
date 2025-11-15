"""
Credit and Risk Scoring API Router
"""

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Optional
from app.models.credit_model import credit_model
from app.models.risk_model import risk_model

router = APIRouter()

class CreditScoringRequest(BaseModel):
    """Request model for credit scoring"""
    account_age_days: float
    monthly_income: Optional[float] = 50000.0
    total_balance: Optional[float] = 100000.0
    transaction_count_30d: Optional[float] = 10.0
    delinquency_count: Optional[float] = 0.0
    loan_history_count: Optional[float] = 0.0
    avg_transaction_amount: Optional[float] = 10000.0
    credit_utilization: Optional[float] = 0.3
    savings_ratio: Optional[float] = 0.2

class RiskScoringRequest(BaseModel):
    """Request model for risk scoring"""
    # Credit features
    account_age_days: float
    monthly_income: Optional[float] = 50000.0
    total_balance: Optional[float] = 100000.0
    transaction_count_30d: Optional[float] = 10.0
    delinquency_count: Optional[float] = 0.0
    loan_history_count: Optional[float] = 0.0
    # Fraud features
    amount: Optional[float] = 0.0
    hour: Optional[float] = 12.0
    day_of_week: Optional[float] = 3.0
    transaction_count_24h: Optional[float] = 0.0
    transaction_count_7d: Optional[float] = 5.0
    avg_amount_7d: Optional[float] = 10000.0
    beneficiary_age_days: Optional[float] = 365.0
    device_risk: Optional[float] = 0.0
    location_risk: Optional[float] = 0.0

@router.post("/credit")
async def predict_credit_score(request: CreditScoringRequest):
    """
    Predict credit score (300-850 range)
    
    Returns credit score, risk category, and score range
    """
    try:
        features = request.dict()
        result = credit_model.predict(features)
        return {
            "success": True,
            "result": result
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Prediction error: {str(e)}")

@router.post("/risk")
async def predict_risk_score(request: RiskScoringRequest):
    """
    Predict overall risk score combining credit and fraud factors
    
    Returns overall risk score, category, and recommendation
    """
    try:
        features = request.dict()
        result = risk_model.predict(features)
        return {
            "success": True,
            "result": result
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Prediction error: {str(e)}")

@router.get("/health")
async def health_check():
    """Health check for scoring models"""
    return {
        "status": "healthy",
        "credit_model_loaded": credit_model.model is not None,
        "risk_model_loaded": True  # Risk model doesn't require file loading
    }

