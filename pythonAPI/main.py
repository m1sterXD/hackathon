from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import joblib
from pydantic import BaseModel
from typing import List

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins="http://localhost:8080",
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE", "OPTIONS"],
    allow_headers=["Content-Type", "Authorization", "Accept"],
    max_age=3600,
)

# model = joblib.load('')
data = {}
class MetricsRequest(BaseModel):
    f1: float
    f2: float
    f3: float
    f4: float
    f5: float

@app.post("/data_ars")
def root(data: MetricsRequest):
    return {
        "received": data.dict()
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
def root(data: List[Features14]):
    return {
        "count": len(data),
        "first": data[0].dict()
    }

@app.post("/predict")
def predict(data: dict):
    return {"prediction": model.predict([data["features"]])[0]}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)