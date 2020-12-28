import requests

for i in range(0,300):
    for j in range(0,5):
        requests.get("http://localhost:20008/DiffAnalysis?id={}&diff={}&speed=-1.0".format(i,j))