from array import array
import numpy as np
import joblib
from pythonAPI.main import predict

model = joblib.load('model_ridge.pkl')
array = np.array([1.1, 1.2, 1.2, 1.4, 1.5]).reshape(1, -1)
predict = model.predict(array)

print(predict)
