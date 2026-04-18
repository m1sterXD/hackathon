import pandas as pd
import numpy as np
import joblib
from typing import List, Dict, Any

# Загружаем модели
_reg = joblib.load("model_ridge.pkl")
_fa = joblib.load("model_efa.pkl")
_scaler = joblib.load("model_scaler.pkl")

FACTOR_LABELS = {
    "F1": "Исследовательская активность",
    "F2": "Качество приёма и престиж",
    "F3": "Международность",
    "F4": "Финансовые ресурсы",
    "F5": "Академическая глубина"
}

# Статистики для нормализации факторов
FACTOR_MEAN = np.array([-0.02, 0.01, -0.01, 0.00, 0.01])
FACTOR_STD = np.array([0.98, 1.02, 0.97, 0.99, 1.01])


USER_COLS_ORDER = [
    "ege_budget", "ege_paid", "ege_min",
    "vsosh_per100", "olymp_per100",
    "rnd_revenue_share",
    "wos_pubs_per100_2021", "grants_per_100_fac",
    "master_share", "foreign_faculty_share",
    "revenue_per_faculty", "modern_equipment_share",
    "postgrad_per_100", "doct_share",
]

def _fa_transform(fa, X):
    """Ручная трансформация через матрицу нагрузок."""
    loadings = fa.loadings_  # shape: (41, 5)
    return X @ loadings @ np.linalg.pinv(loadings.T @ loadings)



def predict(values: List[float]) -> Dict[str, Any]:
    mean     = _scaler["mean"]
    std      = _scaler["std"]
    cols     = _scaler["cols"]       # 41 EFA-колонка
    log_cols = set(_scaler["log_cols"])
    score_min = _scaler["score_min"]
    score_max = _scaler["score_max"]

    # Вектор 41 признака, все = 0 (медиана)
    vec = pd.Series(0.0, index=cols)
    filled_with_median = []
    filled_count = 0

    for i, col in enumerate(USER_COLS_ORDER):
        if i < len(values) and values[i] is not None and values[i] != 0:
            val = float(values[i])          # ← исправлено: i не I
            if col in log_cols:
                val = np.log1p(max(val, 0))
            if col in vec.index:
                vec[col] = (val - mean[col]) / std[col]
            filled_count += 1
        else:
            filled_with_median.append(col)

    # EFA → F1-F5
    factor_scores = _fa_transform(_fa, vec.values.reshape(1, -1))[0]

    # Ridge → Forbes балл
    predicted_score = float(_reg.predict(factor_scores.reshape(1, -1))[0])
    predicted_score = float(np.clip(predicted_score, score_min, score_max))

    # Нормализация 0-100
    score_normalized = round(
        (predicted_score - score_min) / (score_max - score_min) * 100, 1
    )

    # Факторы → шкала 1-10 (по распределению 723 вузов)
    f_bounds = _scaler["f_bounds"]
    factors_result = {}
    for i, fn in enumerate(["F1", "F2", "F3", "F4", "F5"]):
        b = f_bounds[fn]
        norm = (factor_scores[i] - b["min"]) / (b["max"] - b["min"]) * 9 + 1
        norm = round(float(np.clip(norm, 1.0, 10.0)), 2)
        factors_result[fn] = {
            "score":      round(float(factor_scores[i]), 3),
            "normalized": norm,
            "label":      FACTOR_LABELS[fn]
        }

    # Ранг среди 723 вузов
    rank = int((np.array(_scaler["y_extended"]) > score_normalized).sum()) + 1

    return {
        "score_normalized": score_normalized,
        "rank":             rank,
        "factors":          factors_result,
        "data_quality": {
            "filled_count":       filled_count,
            "median_count":       len(filled_with_median),
            "filled_with_median": filled_with_median
        }
    }

if __name__ == "__main__":
    bgu_values = [
        80, 72, 54,  # ege_budget, ege_paid, ege_min
        0.5, 0.3,  # vsosh_per100, olymp_per100
        18,  # rnd_revenue_share
        2.5, 1.2,  # wos_pubs_per100_2021, grants_per_100_fac
        28,  # master_share
        12,  # foreign_faculty_share
        510,  # revenue_per_faculty
        45,  # modern_equipment_share
        3.2,  # postgrad_per_100
        45  # doct_share
    ]
    result = predict(bgu_values)
    print(result)