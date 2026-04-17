import joblib
import numpy as np
import pandas as pd

# ─── Загрузка модели ──────────────────────────────────────────────────────────
MODEL  = joblib.load("model_ridge_direct.pkl")
SCALER = joblib.load("model_scaler_direct.pkl")

# ─── Функция предсказания ─────────────────────────────────────────────────────
def predict(user_input: dict) -> dict:
    cols     = SCALER["cols"]
    log_cols = set(SCALER["log_cols"])
    mean_    = SCALER["mean"]
    std_     = SCALER["std"]

    vec = pd.Series(0.0, index=cols)
    filled_with_median = []

    for col in cols:
        val = user_input.get(col)
        if val is not None:
            if col in log_cols:
                val = np.log1p(max(float(val), 0))
            vec[col] = (float(val) - mean_[col]) / std_[col]
        else:
            filled_with_median.append(col)

    pred      = float(MODEL.predict(vec.values.reshape(1, -1))[0])
    pred      = float(np.clip(pred, SCALER["score_min"] - 5, SCALER["score_max"]))
    s_min, s_max = SCALER["score_min"], SCALER["score_max"]
    normalized   = float(np.clip((pred - s_min) / (s_max - s_min) * 99 + 1, 1, 100))
    est_rank     = int((np.array(SCALER["y_train"]) > pred).sum()) + 1

    return {
        "predicted_score":    round(pred, 2),
        "score_normalized":   round(normalized, 1),
        "est_rank_in_top100": est_rank,
        "filled_with_median": filled_with_median,
    }

# ─── Вывод результата ─────────────────────────────────────────────────────────
def print_result(name: str, result: dict):
    rank  = result["est_rank_in_top100"]
    score = result["score_normalized"]
    raw   = result["predicted_score"]
    imputed = result["filled_with_median"]

    bar_filled = int(score / 5)
    bar = "█" * bar_filled + "░" * (20 - bar_filled)

    print(f"\n{'═'*55}")
    print(f"  ВУЗ: {name}")
    print(f"{'─'*55}")
    print(f"  Балл (1–100):      {score:>6.1f}  [{bar}]")
    print(f"  Forbes raw score:  {raw:>6.2f}")
    print(f"  Место в Forbes-100: #{rank}")
    if rank <= 10:
        print(f"  → Топ-10 Forbes!")
    elif rank <= 25:
        print(f"  → Уровень топ-25")
    elif rank <= 50:
        print(f"  → Уровень топ-50")
    elif rank <= 100:
        print(f"  → Входит в Forbes-100")
    else:
        print(f"  → За пределами Forbes-100")
    if imputed:
        print(f"\n  ⚠ Заменено медианой ({len(imputed)} показателей):")
        for col in imputed:
            print(f"      • {col}")
    print(f"{'═'*55}")


# ─── Тестовые вузы ───────────────────────────────────────────────────────────
if __name__ == "__main__":

    universities = [
        ("БГУ", {
            "ege_budget": 80, "ege_paid": 72, "ege_min": 54,
            "master_share": 28, "postgrad_per_100": 3.2,
            "doct_share": 45, "modern_equipment_share": 45,
            "rnd_revenue_share": 18,
        }),
        ("БГМУ", {
            "ege_budget": 90, "ege_paid": 82, "ege_min": 76,
            "master_share": 20, "postgrad_per_100": 2.8,
            "doct_share": 50, "rnd_revenue_share": 8,
            "modern_equipment_share": 50,
        }),
        ("БНТУ", {
            "ege_budget": 70, "ege_paid": 63, "ege_min": 47,
            "master_share": 22, "postgrad_per_100": 2.1,
            "doct_share": 38, "modern_equipment_share": 38,
            "rnd_revenue_share": 12,
        }),
        ("Идеальный вуз", {
            "ege_budget": 98, "ege_paid": 95, "ege_min": 85,
            "vsosh_per100": 4.0, "olymp_per100": 8.0,
            "rnd_revenue_share": 50, "wos_pubs_per100_2021": 150,
            "grants_per_100_fac": 40, "master_share": 45,
            "foreign_faculty_share": 15, "revenue_per_faculty": 15000,
            "modern_equipment_share": 90, "postgrad_per_100": 12,
            "doct_share": 28,
        }),
    ]

    print("\nПАЙПЛАЙН ПРЕДСКАЗАНИЯ FORBES RANKING")
    print("Модель: Ridge (alpha=50) | LOO R²=0.708 | 14 признаков")

    for name, inputs in universities:
        result = predict(inputs)
        print_result(name, result)
