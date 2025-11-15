"""
Health Check Router
"""

from fastapi import APIRouter

router = APIRouter()

@router.get("/")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "ML Models Service",
        "version": "1.0.0"
    }

@router.get("/ready")
async def readiness_check():
    """Readiness check endpoint"""
    return {
        "status": "ready",
        "models": {
            "fraud_detection": "available",
            "credit_scoring": "available",
            "risk_scoring": "available"
        }
    }

