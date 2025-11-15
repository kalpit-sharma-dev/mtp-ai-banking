"""
Configuration management
"""

from pydantic_settings import BaseSettings
from typing import Optional

class Settings(BaseSettings):
    # Server settings
    HOST: str = "0.0.0.0"
    PORT: int = 9000
    DEBUG: bool = False
    
    # Model paths
    FRAUD_MODEL_PATH: str = "models/fraud_detection_model.pkl"
    CREDIT_MODEL_PATH: str = "models/credit_scoring_model.pkl"
    RISK_MODEL_PATH: str = "models/risk_scoring_model.pkl"
    
    # Model settings
    FRAUD_THRESHOLD: float = 0.5
    CREDIT_SCORE_MIN: int = 300
    CREDIT_SCORE_MAX: int = 850
    
    # API settings
    API_KEY: Optional[str] = None
    
    class Config:
        env_file = ".env"
        case_sensitive = False

settings = Settings()

