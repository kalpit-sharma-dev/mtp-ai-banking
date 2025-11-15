"""
Model Training Script
Trains fraud detection and credit scoring models
"""

import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestRegressor
import xgboost as xgb
import joblib
import os

def generate_synthetic_data(n_samples=10000):
    """Generate synthetic training data"""
    np.random.seed(42)
    
    # Fraud detection features
    fraud_data = {
        'amount': np.random.lognormal(10, 1, n_samples),
        'hour': np.random.randint(0, 24, n_samples),
        'day_of_week': np.random.randint(0, 7, n_samples),
        'transaction_count_24h': np.random.poisson(2, n_samples),
        'transaction_count_7d': np.random.poisson(10, n_samples),
        'avg_amount_7d': np.random.lognormal(9.5, 0.8, n_samples),
        'beneficiary_age_days': np.random.exponential(30, n_samples),
        'device_risk': np.random.beta(2, 5, n_samples),
        'location_risk': np.random.beta(2, 5, n_samples),
        'user_account_age_days': np.random.exponential(365, n_samples),
        'user_balance': np.random.lognormal(11, 1, n_samples),
        'is_new_beneficiary': np.random.binomial(1, 0.2, n_samples),
        'is_unusual_hour': np.random.binomial(1, 0.1, n_samples),
        'amount_vs_avg_ratio': np.random.lognormal(0, 0.5, n_samples),
        'velocity_score': np.random.beta(2, 5, n_samples)
    }
    
    # Generate fraud labels (higher probability for suspicious patterns)
    fraud_labels = []
    for i in range(n_samples):
        prob = 0.1  # Base probability
        if fraud_data['amount'][i] > 100000:
            prob += 0.3
        if fraud_data['beneficiary_age_days'][i] < 7:
            prob += 0.2
        if fraud_data['transaction_count_24h'][i] > 5:
            prob += 0.2
        if fraud_data['device_risk'][i] > 0.5:
            prob += 0.2
        prob = min(prob, 0.95)
        fraud_labels.append(np.random.binomial(1, prob))
    
    # Credit scoring features
    credit_data = {
        'account_age_days': np.random.exponential(365, n_samples),
        'monthly_income': np.random.lognormal(10.5, 0.5, n_samples),
        'total_balance': np.random.lognormal(11, 1, n_samples),
        'transaction_count_30d': np.random.poisson(20, n_samples),
        'delinquency_count': np.random.poisson(0.5, n_samples),
        'loan_history_count': np.random.poisson(2, n_samples),
        'avg_transaction_amount': np.random.lognormal(9, 0.8, n_samples),
        'credit_utilization': np.random.beta(2, 3, n_samples),
        'savings_ratio': np.random.beta(3, 2, n_samples)
    }
    
    # Generate credit scores (300-850 range)
    credit_scores = []
    for i in range(n_samples):
        score = 600  # Base score
        score += min(credit_data['account_age_days'][i] / 10, 50)
        score += min(credit_data['monthly_income'][i] / 1000, 100)
        score -= credit_data['delinquency_count'][i] * 20
        score += credit_data['loan_history_count'][i] * 10
        score = max(300, min(850, score + np.random.normal(0, 30)))
        credit_scores.append(score)
    
    return pd.DataFrame(fraud_data), np.array(fraud_labels), pd.DataFrame(credit_data), np.array(credit_scores)

def train_fraud_model():
    """Train fraud detection model"""
    print("Generating synthetic fraud data...")
    fraud_df, fraud_labels, _, _ = generate_synthetic_data(10000)
    
    print("Training fraud detection model...")
    X_train, X_test, y_train, y_test = train_test_split(
        fraud_df, fraud_labels, test_size=0.2, random_state=42
    )
    
    model = xgb.XGBClassifier(
        n_estimators=100,
        max_depth=5,
        learning_rate=0.1,
        random_state=42
    )
    
    model.fit(X_train, y_train)
    
    # Evaluate
    train_score = model.score(X_train, y_train)
    test_score = model.score(X_test, y_test)
    print(f"Fraud Model - Train Score: {train_score:.4f}, Test Score: {test_score:.4f}")
    
    # Save model
    os.makedirs("models", exist_ok=True)
    model_path = "models/fraud_detection_model.pkl"
    joblib.dump(model, model_path)
    print(f"Saved fraud model to {model_path}")
    
    return model

def train_credit_model():
    """Train credit scoring model"""
    print("Generating synthetic credit data...")
    _, _, credit_df, credit_scores = generate_synthetic_data(10000)
    
    print("Training credit scoring model...")
    X_train, X_test, y_train, y_test = train_test_split(
        credit_df, credit_scores, test_size=0.2, random_state=42
    )
    
    model = RandomForestRegressor(
        n_estimators=100,
        max_depth=10,
        random_state=42
    )
    
    model.fit(X_train, y_train)
    
    # Evaluate
    train_score = model.score(X_train, y_train)
    test_score = model.score(X_test, y_test)
    print(f"Credit Model - Train R²: {train_score:.4f}, Test R²: {test_score:.4f}")
    
    # Save model
    os.makedirs("models", exist_ok=True)
    model_path = "models/credit_scoring_model.pkl"
    joblib.dump(model, model_path)
    print(f"Saved credit model to {model_path}")
    
    return model

if __name__ == "__main__":
    print("=" * 50)
    print("Training ML Models")
    print("=" * 50)
    
    # Train fraud model
    print("\n1. Training Fraud Detection Model")
    print("-" * 50)
    train_fraud_model()
    
    # Train credit model
    print("\n2. Training Credit Scoring Model")
    print("-" * 50)
    train_credit_model()
    
    print("\n" + "=" * 50)
    print("Training Complete!")
    print("=" * 50)

