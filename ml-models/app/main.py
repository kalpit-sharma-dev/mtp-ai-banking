"""
ML Models Service - Layer 4
Serves fraud detection and scoring models via REST API
"""

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
from app.config import settings
from app.routers import fraud, scoring, health

app = FastAPI(
    title="ML Models Service",
    description="Fraud Detection and Scoring Models API",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(health.router, prefix="/health", tags=["health"])
app.include_router(fraud.router, prefix="/api/v1/fraud", tags=["fraud"])
app.include_router(scoring.router, prefix="/api/v1/scoring", tags=["scoring"])

@app.get("/")
async def root():
    return {
        "service": "ML Models Service",
        "version": "1.0.0",
        "models": ["fraud_detection", "credit_scoring", "risk_scoring"]
    }

@app.get("/health")
async def health():
    """Health check endpoint (direct)"""
    return {
        "status": "healthy",
        "service": "ML Models Service",
        "version": "1.0.0"
    }

if __name__ == "__main__":
    uvicorn.run(
        "app.main:app",
        host=settings.HOST,
        port=settings.PORT,
        reload=settings.DEBUG
    )

