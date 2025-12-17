import joblib
import os
from pathlib import Path

def summarizeReadme(data):
    BASE_DIR = Path(__file__).resolve().parent.parent
    MODEL_PATH = os.path.join(BASE_DIR, 'ml', 'summarizer_model.joblib')
    return None