"""
Fraud Detection Model
Uses XGBoost for fraud detection
"""

import numpy as np
import joblib
import os
from typing import Dict, List, Optional
from app.config import settings

class FraudDetectionModel:
    """Fraud detection model using XGBoost"""
    
    def __init__(self, model_path: Optional[str] = None):
        self.model_path = model_path or settings.FRAUD_MODEL_PATH
        self.model = None
        self.feature_names = [
            'amount', 'hour', 'day_of_week', 'transaction_count_24h',
            'transaction_count_7d', 'avg_amount_7d', 'beneficiary_age_days',
            'device_risk', 'location_risk', 'user_account_age_days',
            'user_balance', 'is_new_beneficiary', 'is_unusual_hour',
            'amount_vs_avg_ratio', 'velocity_score'
        ]
        self.load_model()
    
    def load_model(self):
        """Load the trained model"""
        if os.path.exists(self.model_path):
            try:
                self.model = joblib.load(self.model_path)
                print(f"Loaded fraud detection model from {self.model_path}")
            except Exception as e:
                print(f"Error loading model: {e}. Using mock model.")
                self.model = None
        else:
            print(f"Model file not found at {self.model_path}. Using mock model.")
            self.model = None
    
    def predict(self, features: Dict[str, float]) -> Dict[str, float]:
        """
        Predict fraud probability
        
        Args:
            features: Dictionary of feature values
            
        Returns:
            Dictionary with fraud_score and risk_level
        """
        if self.model is None:
            # Mock prediction if model not loaded
            return self._mock_predict(features)
        
        # Prepare feature vector
        feature_vector = self._prepare_features(features)
        
        # Predict
        fraud_probability = self.model.predict_proba([feature_vector])[0][1]
        
        # Determine risk level
        risk_level = self._get_risk_level(fraud_probability)
        
        return {
            "fraud_score": float(fraud_probability),
            "risk_level": risk_level,
            "is_fraud": fraud_probability >= settings.FRAUD_THRESHOLD
        }
    
    def _prepare_features(self, features: Dict[str, float]) -> List[float]:
        """Prepare feature vector from input dictionary"""
        feature_vector = []
        for feature_name in self.feature_names:
            value = features.get(feature_name, 0.0)
            feature_vector.append(float(value))
        return feature_vector
    
    def _get_risk_level(self, score: float) -> str:
        """Get risk level from fraud score"""
        if score >= 0.7:
            return "HIGH"
        elif score >= 0.4:
            return "MEDIUM"
        else:
            return "LOW"
    
    def _mock_predict(self, features: Dict[str, float]) -> Dict[str, float]:
        """Mock prediction when model is not available"""
        amount = features.get('amount', 0.0)
        beneficiary_age = features.get('beneficiary_age_days', 365.0)
        txn_count_24h = features.get('transaction_count_24h', 0.0)
        
        # Simple rule-based scoring
        score = 0.0
        
        # Amount factor
        if amount > 200000:
            score += 0.4
        elif amount > 100000:
            score += 0.2
        elif amount > 50000:
            score += 0.1
        
        # New beneficiary
        if beneficiary_age < 7:
            score += 0.3
        
        # Velocity
        if txn_count_24h > 10:
            score += 0.3
        elif txn_count_24h > 5:
            score += 0.15
        
        # Device/location risk
        score += features.get('device_risk', 0.0) * 0.1
        score += features.get('location_risk', 0.0) * 0.1
        
        if score > 1.0:
            score = 1.0
        
        return {
            "fraud_score": float(score),
            "risk_level": self._get_risk_level(score),
            "is_fraud": score >= settings.FRAUD_THRESHOLD
        }

# Global model instance
fraud_model = FraudDetectionModel()

