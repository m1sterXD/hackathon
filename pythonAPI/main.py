from fastapi import FastAPI
from fastapi import Body
from fastapi.middleware.cors import CORSMiddleware
import joblib
from pydantic import BaseModel
from typing import List
from predictor_direct import predict


app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins="http://localhost:8080",
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE", "OPTIONS"],
    allow_headers=["Content-Type", "Authorization", "Accept"],
    max_age=3600,
)

model = joblib.load('model_ridge.pkl')
data = {}
class MetricsRequest(BaseModel):
    f1: float
    f2: float
    f3: float
    f4: float
    f5: float

@app.post("/data_ars")
def root(data: MetricsRequest):
    features = [
        data.f1,
        data.f2,
        data.f3,
        data.f4,
        data.f5,
    ]

    result = model.predict([features])[0]

    return {
        "score": float(result)
    }


class Features14(BaseModel):
    f1: float
    f2: float
    f3: float
    f4: float
    f5: float
    f6: float
    f7: float
    f8: float
    f9: float
    f10: float
    f11: float
    f12: float
    f13: float
    f14: float

@app.post("/data_ars_forteen")
def root(data: List[Features14] = Body(...)):
    item = data[0]

    features = [
        item.f1, item.f2, item.f3, item.f4,
        item.f5, item.f6, item.f7, item.f8,
        item.f9, item.f10, item.f11, item.f12,
        item.f13, item.f14
    ]

    result = predict(features)

    return {
        "score": float(result)
    }

@app.post("/predict")
def predict(data: dict):
    return {"prediction": model.predict([data["features"]])[0]}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)