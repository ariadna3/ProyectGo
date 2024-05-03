
import requests
import json
import datetime
import sys
import pymongo

jsonQueryFile = open(sys.argv[1])
data = json.load(jsonQueryFile)

#datos para la conexion con SQL
SERVER = data["server"]
DATABASE = data["database"]
USER = data["user"]
PASSWORD = data["password"]

#datos para la conexion con Mongo
MONGO_URL = data["mongo_url"]
MONGO_DB = data["mongo_db"]
MONGO_USER = data["mongo_user"]
MONGO_PASSWORD = data["mongo_password"]

#datos para la conexion con las apis
TOKEN = data["token"]
URLS = data["urls"]

#querys
QUERYS = data["querys"]




def main():
    mydb = connect_mongo(url=MONGO_URL, database=MONGO_DB, user=MONGO_USER, password=MONGO_PASSWORD)
    for query in QUERYS:
        package_to_send = get_package_from_mongo(mydb, query["Query"], query["Tabla"])
        path: str = query["Path"]

        response = send_package_to_api(url=URLS[query["Server"]], path=path,  package=package_to_send, token=TOKEN)

        print(response.text)







def connect_mongo(url: str, database: str, user: str, password: str) -> pymongo.database.Database:
    myclient = pymongo.MongoClient(url)
    return myclient[database]

def connect_sql(server: str, database: str, user: str, password: str):
    pass

def get_data_mongo(db, query: str, collection: str) -> list:
    array = [d for d in db[collection].find(query)]
    #Remove _id from the response
    array = [{k: v for k, v in d.items() if k != "_id"} for d in array]
    #put the keys in lowercase
    array = [{k.lower(): v for k, v in d.items()} for d in array]
    #If type of a value is datetime, convert it to string
    usanGerente10001 = [1024, 1814, 1291, 1953]
    usanGerente10002 = [1368, 1813, 76, 1965, 1774, 1773, 1989, 1935, 1803, 1714, 1356, 1839, 1677, 1902]
    for d in array:
        for k, v in d.items():
            if isinstance(v, datetime.datetime):
                d[k] = str(v)
        if d["legajo"] in usanGerente10002:
            d["gerente"] = 10002
        if d["legajo"] in usanGerente10001:
            d["gerente"] = 10001 
            
    return array

def get_package_from_mongo(db, query: str, collection: str) -> dict:
    return {"Paquete": get_data_mongo(db, query, collection)}

def get_data_sql(query: str):
    pass

def send_package_to_api(url: str,path: str, package: dict, token: str) -> requests.Response:
    url_complete = url + path
    headers = {'Authorization': f'Bearer {TOKEN}'}
    headers = {
        'Authorization': 'Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjZjZTExYWVjZjllYjE0MDI0YTQ0YmJmZDFiY2Y4YjMyYTEyMjg3ZmEiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiI2OTk2ODQzMDUyMzktNDRpZDFlNzBhM2Y1aG40cGpiMnVrNWFkMmRjcTQ2ZmkuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJhdWQiOiI2OTk2ODQzMDUyMzktNDRpZDFlNzBhM2Y1aG40cGpiMnVrNWFkMmRjcTQ2ZmkuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJzdWIiOiIxMDAyODE5NTY2NDU0NTIzNzQ2ODUiLCJoZCI6Iml0cGF0YWdvbmlhLmNvbSIsImVtYWlsIjoiYWdvbnphbGV6QGl0cGF0YWdvbmlhLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJuYmYiOjE3MTM3MzM0ODYsIm5hbWUiOiJBcmlhZG5hIEdvbnphbGV6IiwicGljdHVyZSI6Imh0dHBzOi8vbGgzLmdvb2dsZXVzZXJjb250ZW50LmNvbS9hL0FDZzhvY0tHRFJWV2dZYVM0amtJNG1rMkJHb1BlaTdqcmxjem56Vy1DWERCV3B4aGxNQmR4dz1zOTYtYyIsImdpdmVuX25hbWUiOiJBcmlhZG5hIiwiZmFtaWx5X25hbWUiOiJHb256YWxleiIsImlhdCI6MTcxMzczMzc4NiwiZXhwIjoxNzEzNzM3Mzg2LCJqdGkiOiI3YzI2MDZlZDA3ZTFmMmQ0NTM2MjkyNjhmNDk5YjBjMWQ2N2YyMTQxIn0.Rp3fTPjB1A33a8u-AjVj9492YdpEU2y9LND0jDBwLtrR5aDxbbCwoWXmK4Ssc7yEOc8-wtZNDGZQJMDRVAYjf8Xcx3V2WzCc07_8sLeMBOL001uLgVitTbCxsnqf1rUb-K4eVPs3KkCeF1_PAbBawHRNnGsQrNZiTICL5hYBadlg4gwTcuu-XY8vG85FuVgxfT8IQcV8m_GGEyQ5AZRfb6I2h0Df-lFX56NKupGpI1RsnGtnlJJ9GeFwc3LVueauHr7gYuQbpqApCRZ-P-FbTBshCEiNWEpSQvCf-Liw7v5S8rElAYsLo7TFEzt-orcfZcLhg7Kbv8ABiwUgWuW2GA'
    }

    print(headers)

    return requests.post(url=url_complete, json=package, headers=headers)

if __name__ == "__main__":
    main()