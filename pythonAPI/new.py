import joblib
import numpy as np
import pandas as pd

# ─── Загрузка модели ──────────────────────────────────────────────────────────
MODEL = joblib.load("model_ridge_direct.pkl")
SCALER = joblib.load("model_scaler_direct.pkl")


# ─── Функция предсказания из списка (14 параметров) ─────────────────────────
def predict_score(values: list) -> float:
    """
    Принимает список из 14 параметров, возвращает предсказанный score (1-100)

    Порядок параметров (14 штук):
    0. ege_budget
    1. ege_paid
    2. ege_min
    3. master_share
    4. postgrad_per_100
    5. doct_share
    6. modern_equipment_share
    7. rnd_revenue_share
    8. vsosh_per100
    9. olymp_per100
    10. wos_pubs_per100_2021
    11. grants_per_100_fac
    12. foreign_faculty_share
    13. revenue_per_faculty
    """
    cols = SCALER["cols"]
    log_cols = set(SCALER["log_cols"])
    mean_ = SCALER["mean"]
    std_ = SCALER["std"]

    # Проверяем длину
    if len(values) != len(cols):
        raise ValueError(f"Ожидается {len(cols)} параметров, получено {len(values)}")

    # Создаем вектор признаков
    vec = pd.Series(0.0, index=cols)

    for i, col in enumerate(cols):
        val = values[i]

        # Если значение есть и не 0 (0 = нет данных)
        if val is not None and val != 0:
            if col in log_cols:
                val = np.log1p(max(float(val), 0))
            vec[col] = (float(val) - mean_[col]) / std_[col]
        # else: оставляем 0 (замена медианой)

    # Предсказание
    pred = float(MODEL.predict(vec.values.reshape(1, -1))[0])
    pred = float(np.clip(pred, SCALER["score_min"] - 5, SCALER["score_max"]))

    # Нормализация в диапазон 1-100
    s_min, s_max = SCALER["score_min"], SCALER["score_max"]
    normalized = float(np.clip((pred - s_min) / (s_max - s_min) * 99 + 1, 1, 100))

    return round(normalized, 1)

# ГЛАВНАЯ ФУНКЦИЯ ДЛЯ ЗАПУСКА
# ────────────────────────────────────────────────────────────────────────────
if __name__ == "__main__":

    # ─── Тестовые данные (списки из 14 параметров) ────────────────────────

    # БГУ (некоторые параметры = 0 - нет данных)
    bgu_params = [80, 72, 54, 28, 3.2, 45, 45, 18, 0, 0, 0, 0, 0, 0]

    # БГМУ
    bgmu_params = [90, 82, 76, 20, 2.8, 50, 50, 8, 0, 0, 0, 0, 0, 0]

    # БНТУ
    bntu_params = [70, 63, 47, 22, 2.1, 38, 38, 12, 0, 0, 0, 0, 0, 0]

    # Идеальный вуз (все параметры заполнены)
    ideal_params = [98, 95, 85, 45, 12, 28, 90, 50, 4.0, 8.0, 150, 40, 15, 15000]

    # Слабый вуз (низкие параметры)
    weak_params = [50, 45, 40, 10, 0.5, 15, 20, 5, 0, 0, 0, 0, 0, 0]

    # ─── Запуск предсказаний ───────────────────────────────────────────────

    result = predict_score(ideal_params)
    print(result)