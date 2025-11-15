"""
Fraud Detection API Router
"""

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Dict, Optional
from app.models.fraud_model import fraud_model

router = APIRouter()

class FraudPredictionRequest(BaseModel):
    """Request model for fraud prediction"""
    amount: float
    hour: Optional[float] = 12.0
    day_of_week: Optional[float] = 3.0
    transaction_count_24h: Optional[float] = 0.0
    transaction_count_7d: Optional[float] = 0.0
    avg_amount_7d: Optional[float] = 10000.0
    beneficiary_age_days: Optional[float] = 365.0
    device_risk: Optional[float] = 0.0
    location_risk: Optional[float] = 0.0
    user_account_age_days: Optional[float] = 365.0
    user_balance: Optional[float] = 100000.0

@router.post("/predict")
async def predict_fraud(request: FraudPredictionRequest):
    """
    Predict fraud probability for a transaction
    
    Returns fraud score (0.0 to 1.0) and risk level
    """
    try:
        features = request.dict()
        result = fraud_model.predict(features)
        return {
            "success": True,
            "result": result
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Prediction error: {str(e)}")

@router.get("/health")
async def health_check():
    """Health check for fraud model"""
    return {
        "status": "healthy",
        "model_loaded": fraud_model.model is not None
    }

