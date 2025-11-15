"""
Risk Scoring Model
Combines credit and fraud scores for overall risk assessment
"""

from typing import Dict
from app.models.fraud_model import fraud_model
from app.models.credit_model import credit_model

class RiskScoringModel:
    """Overall risk scoring model combining multiple factors"""
    
    def predict(self, features: Dict) -> Dict[str, any]:
        """
        Predict overall risk score
        
        Args:
            features: Dictionary containing both credit and fraud features
            
        Returns:
            Dictionary with overall_risk_score and components
        """
        # Extract credit features
        credit_features = {
            'account_age_days': features.get('account_age_days', 365.0),
            'monthly_income': features.get('monthly_income', 50000.0),
            'total_balance': features.get('total_balance', 100000.0),
            'transaction_count_30d': features.get('transaction_count_30d', 10.0),
            'delinquency_count': features.get('delinquency_count', 0.0),
            'loan_history_count': features.get('loan_history_count', 0.0),
            'avg_transaction_amount': features.get('avg_transaction_amount', 10000.0),
            'credit_utilization': features.get('credit_utilization', 0.3),
            'savings_ratio': features.get('savings_ratio', 0.2)
        }
        
        # Extract fraud features
        fraud_features = {
            'amount': features.get('amount', 0.0),
            'hour': features.get('hour', 12.0),
            'day_of_week': features.get('day_of_week', 3.0),
            'transaction_count_24h': features.get('transaction_count_24h', 0.0),
            'transaction_count_7d': features.get('transaction_count_7d', 5.0),
            'avg_amount_7d': features.get('avg_amount_7d', 10000.0),
            'beneficiary_age_days': features.get('beneficiary_age_days', 365.0),
            'device_risk': features.get('device_risk', 0.0),
            'location_risk': features.get('location_risk', 0.0),
            'user_account_age_days': features.get('account_age_days', 365.0),
            'user_balance': features.get('total_balance', 100000.0),
            'is_new_beneficiary': 1.0 if features.get('beneficiary_age_days', 365.0) < 7 else 0.0,
            'is_unusual_hour': 1.0 if features.get('hour', 12.0) < 6 or features.get('hour', 12.0) > 23 else 0.0,
            'amount_vs_avg_ratio': features.get('amount', 0.0) / max(features.get('avg_amount_7d', 1.0), 1.0),
            'velocity_score': min(features.get('transaction_count_24h', 0.0) / 10.0, 1.0)
        }
        
        # Get credit score
        credit_result = credit_model.predict(credit_features)
        credit_score = credit_result['credit_score']
        
        # Convert credit score to risk (inverse: higher score = lower risk)
        credit_risk = 1.0 - (credit_score / 850.0)
        
        # Get fraud score
        fraud_result = fraud_model.predict(fraud_features)
        fraud_risk = fraud_result['fraud_score']
        
        # Calculate amount risk
        amount = features.get('amount', 0.0)
        amount_risk = self._calculate_amount_risk(amount)
        
        # Weighted combination
        overall_risk = (
            credit_risk * 0.4 +    # 40% weight on credit risk
            fraud_risk * 0.4 +     # 40% weight on fraud risk
            amount_risk * 0.2      # 20% weight on amount risk
        )
        
        # Ensure in 0-1 range
        overall_risk = max(0.0, min(1.0, overall_risk))
        
        return {
            "overall_risk_score": float(overall_risk),
            "risk_category": self._get_risk_category(overall_risk),
            "components": {
                "credit_risk": float(credit_risk),
                "fraud_risk": float(fraud_risk),
                "amount_risk": float(amount_risk)
            },
            "credit_score": credit_score,
            "fraud_score": float(fraud_risk),
            "recommendation": self._get_recommendation(overall_risk)
        }
    
    def _calculate_amount_risk(self, amount: float) -> float:
        """Calculate risk based on transaction amount"""
        if amount > 200000:
            return 0.8
        elif amount > 100000:
            return 0.5
        elif amount > 50000:
            return 0.3
        else:
            return 0.1
    
    def _get_risk_category(self, risk: float) -> str:
        """Get risk category from overall risk score"""
        if risk < 0.3:
            return "LOW"
        elif risk < 0.6:
            return "MEDIUM"
        else:
            return "HIGH"
    
    def _get_recommendation(self, risk: float) -> str:
        """Get recommendation based on risk score"""
        if risk > 0.7:
            return "BLOCK"
        elif risk > 0.4:
            return "REVIEW"
        else:
            return "APPROVE"

# Global model instance
risk_model = RiskScoringModel()

