# ML Models Service - Setup Guide

## Quick Setup

### Option 1: Automatic (via start-all.bat)
The `start-all.bat` script will automatically install dependencies before starting the service.

### Option 2: Manual Setup

1. **Navigate to ml-models directory:**
   ```bash
   cd ml-models
   ```

2. **Create virtual environment (recommended):**
   ```bash
   python -m venv venv
   
   # On Windows:
   venv\Scripts\activate
   
   # On Linux/Mac:
   source venv/bin/activate
   ```

3. **Install dependencies:**
   ```bash
   pip install -r requirements.txt
   ```

4. **Run the service:**
   ```bash
   python -m app.main
   # OR
   uvicorn app.main:app --host 0.0.0.0 --port 9000
   ```

## Required Dependencies

The service requires:
- `fastapi` - Web framework
- `uvicorn` - ASGI server
- `pydantic` - Data validation
- `pydantic-settings` - Settings management
- `numpy` - Numerical computing
- `pandas` - Data manipulation
- `scikit-learn` - Machine learning
- `xgboost` - Gradient boosting
- `joblib` - Model serialization
- `python-dotenv` - Environment variables

## Troubleshooting

### ModuleNotFoundError: No module named 'pydantic_settings'

**Solution:**
```bash
cd ml-models
pip install -r requirements.txt
```

### Python not found

Make sure Python 3.8+ is installed and in your PATH:
```bash
python --version
```

### Virtual environment issues

If using a virtual environment, make sure it's activated before installing:
```bash
# Windows
venv\Scripts\activate

# Linux/Mac
source venv/bin/activate
```

## Verification

After installation, verify the service starts:
```bash
cd ml-models
python -m app.main
```

You should see:
```
INFO:     Started server process
INFO:     Waiting for application startup.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://0.0.0.0:9000
```

Then test the health endpoint:
```bash
curl http://localhost:9000/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "ML Models Service",
  "version": "1.0.0"
}
```

