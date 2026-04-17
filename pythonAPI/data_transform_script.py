
import joblib, numpy as np, pandas as pd

scaler = joblib.load("model_scaler.pkl")  # mean, std, cols, log_cols
fa     = joblib.load("model_efa.pkl")     # EFA 41→5
reg    = joblib.load("model_ridge.pkl")   # Ridge 5→score

def predict(user_input: dict) -> dict:
    # Шаг 1-3: сырые → z-score → вектор 41
    vec = pd.Series(0.0, index=scaler["cols"])  # все = медиана
    for col, val in user_input.items():
        if col not in vec.index: continue
        if col in scaler["log_cols"]:
            val = np.log1p(max(val, 0))
        vec[col] = (val - scaler["mean"][col]) / scaler["std"][col]

    # Шаг 4: EFA → F1..F5
    factors = fa.transform(vec.values.reshape(1, -1))[0]

    # Шаг 5: Ridge → score
    score = reg.predict(factors.reshape(1, -1))[0]

    return {
        "F1": round(factors[0], 3),
        "F2": round(factors[1], 3),
        "F3": round(factors[2], 3),
        "F4": round(factors[3], 3),
        "F5": round(factors[4], 3),
        "predicted_score": round(float(score), 2)
    }