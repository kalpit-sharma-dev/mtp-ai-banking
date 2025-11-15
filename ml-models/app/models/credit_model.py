"""
Credit Scoring Model
Uses Random Forest for credit scoring
"""

import numpy as np
import joblib
import os
from typing import Dict, Optional
from app.config import settings

class CreditScoringModel:
    """Credit scoring model using Random Forest"""
    
    def __init__(self, model_path: Optional[str] = None):
        self.model_path = model_path or settings.CREDIT_MODEL_PATH
        self.model = None
        self.feature_names = [
            'account_age_days', 'monthly_income', 'total_balance',
            'transaction_count_30d', 'delinquency_count', 'loan_history_count',
            'avg_transaction_amount', 'credit_utilization', 'savings_ratio'
        ]
        self.load_model()
    
    def load_model(self):
        """Load the trained model"""
        if os.path.exists(self.model_path):
            try:
                self.model = joblib.load(self.model_path)
                print(f"Loaded credit scoring model from {self.model_path}")
            except Exception as e:
                print(f"Error loading model: {e}. Using mock model.")
                self.model = None
        else:
            print(f"Model file not found at {self.model_path}. Using mock model.")
            self.model = None
    
    def predict(self, features: Dict[str, float]) -> Dict[str, any]:
        """
        Predict credit score
        
        Args:
            features: Dictionary of feature values
            
        Returns:
            Dictionary with credit_score and risk_category
        """
        if self.model is None:
            # Mock prediction if model not loaded
            return self._mock_predict(features)
        
        # Prepare feature vector
        feature_vector = self._prepare_features(features)
        
        # Predict credit score (300-850 range)
        credit_score = self.model.predict([feature_vector])[0]
        
        # Ensure score is in valid range
        credit_score = max(settings.CREDIT_SCORE_MIN, min(settings.CREDIT_SCORE_MAX, credit_score))
        
        # Calculate risk category
        risk_category = self._get_risk_category(credit_score)
        score_range = self._get_score_range(credit_score)
        
        return {
            "credit_score": int(credit_score),
            "risk_category": risk_category,
            "score_range": score_range,
            "factors": self._get_factor_analysis(features)
        }
    
    def _prepare_features(self, features: Dict[str, float]) -> list:
        """Prepare feature vector from input dictionary"""
        feature_vector = []
        for feature_name in self.feature_names:
            value = features.get(feature_name, 0.0)
            feature_vector.append(float(value))
        return feature_vector
    
    def _get_risk_category(self, score: float) -> str:
        """Get risk category from credit score"""
        if score >= 750:
            return "LOW"
        elif score >= 700:
            return "MEDIUM_LOW"
        elif score >= 650:
            return "MEDIUM"
        elif score >= 600:
            return "MEDIUM_HIGH"
        else:
            return "HIGH"
    
    def _get_score_range(self, score: float) -> str:
        """Get score range classification"""
        if score >= 750:
            return "EXCELLENT"
        elif score >= 700:
            return "GOOD"
        elif score >= 650:
            return "FAIR"
        elif score >= 600:
            return "POOR"
        else:
            return "VERY_POOR"
    
    def _get_factor_analysis(self, features: Dict[str, float]) -> Dict[str, str]:
        """Analyze factors affecting credit score"""
        factors = {}
        
        account_age = features.get('account_age_days', 0)
        if account_age < 90:
            factors['account_age'] = "New account - negative impact"
        elif account_age > 365:
            factors['account_age'] = "Established account - positive impact"
        
        delinquency = features.get('delinquency_count', 0)
        if delinquency > 0:
            factors['delinquency'] = f"{delinquency} delinquencies - negative impact"
        
        income = features.get('monthly_income', 0)
        if income > 100000:
            factors['income'] = "High income - positive impact"
        elif income < 25000:
            factors['income'] = "Low income - negative impact"
        
        return factors
    
    def _mock_predict(self, features: Dict[str, float]) -> Dict[str, any]:
        """Mock prediction when model is not available"""
        account_age = features.get('account_age_days', 365.0)
        income = features.get('monthly_income', 50000.0)
        balance = features.get('total_balance', 100000.0)
        delinquency = features.get('delinquency_count', 0.0)
        loan_history = features.get('loan_history_count', 0.0)
        
        # Base score
        score = 600.0
        
        # Account age factor
        if account_age > 365:
            score += 50
        elif account_age > 180:
            score += 30
        elif account_age > 90:
            score += 15
        
        # Income factor
        if income > 100000:
            score += 100
        elif income > 50000:
            score += 60
        elif income > 25000:
            score += 30
        
        # Delinquency penalty
        score -= delinquency * 20
        
        # Loan history bonus
        if loan_history > 0:
            score += 30
        
        # Balance factor
        if balance > 100000:
            score += 50
        elif balance > 50000:
            score += 30
        
        # Cap at valid range
        score = max(settings.CREDIT_SCORE_MIN, min(settings.CREDIT_SCORE_MAX, score))
        
        return {
            "credit_score": int(score),
            "risk_category": self._get_risk_category(score),
            "score_range": self._get_score_range(score),
            "factors": self._get_factor_analysis(features)
        }

# Global model instance
credit_model = CreditScoringModel()

