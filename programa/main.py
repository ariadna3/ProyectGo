import pyodbc
import requests
import json
import datetime
import sys

jsonQueryFile = open(sys.argv[1])
data = json.load(jsonQueryFile)

#datos para la conexion con SQL
SERVER = data["server"]
DATABASE = data["database"]
USER = data["user"]
PASSWORD = data["password"]

#datos para la conexion con las apis
TOKEN = data["token"]
URLS = data["urls"]

#querys de SQL
QUERYS = data["querys"]

def main():
    packageArray = []
    package = {"Paquete": packageArray}
    inicio = datetime.datetime.now()
    contador_for = 0
    datos_for = []
    mensaje = ""
    try:
        conexionLogin = f'Trusted_Connection=yes;'
        if PASSWORD != "" and USER != "":
            conexionLogin = f'Password={PASSWORD};User={USER};'
        conn = pyodbc.connect("DRIVER={SQL Server Native Client 11.0};"
                              f'SERVER={SERVER};'
                              f'DATABASE={DATABASE};{conexionLogin}'
                              )
        mycursor = conn.cursor()
        headers = {'Authorization': f'Bearer {TOKEN}'}
        print('Conexión establecida con SQL Server')
        for query in QUERYS:
            mycursor.execute(query["Query"])
            for row in mycursor.fetchall():
                # The API endpoint
                url = URLS[query["Server"]] + query["Path"]
                datos_for.append({
                    "server": query["Server"],
                    "tabla": query["Tabla"],
                    "estado": "iniciado"
                })
                ultimo_dato = {}
                # Adding a payload
                response = ""
                try:
                    for data in row:
                        package["Paquete"].append(data)
                    payload = json.loads(package)
                    ultimo_dato = payload[0]

                    # A get request to the API
                    response = requests.post(url, json=payload[0], headers=headers)

                    # Print the response
                    print(response)
                    response_json = response.json()

                    for i in response_json:
                        print(i, "\n")
                    contador_for += 1
                    datos_for[-1]["estado"] = "finalizado correctamente"
                except:
                    datos_for[-1]["estado"] = "error al subir datos Response: " + str(response) + ".Ultimo dato subido: " + str(ultimo_dato)
            conn.close()
    except pyodbc.Error as e:
        mensaje = 'Error al conectarse a SQL Server:' + str(e)
    except Exception as e:
        mensaje = str(e)
    finally:
        fin = datetime.datetime.now()
        duracion = fin - inicio
        escribir_log(inicio, fin, duracion, contador_for, datos_for, mensaje)

if __name__ == "__main__":
    main()

def escribir_log(inicio, fin, duracion, contador_for, datos_for, mensaje):
    with open("log.txt", "a") as archivo:
        archivo.write(str(datetime.datetime.now()) + "\n")
        if mensaje != "":
            archivo.write(f"Mensaje de error: {mensaje}\n")
        archivo.write(f"Inicio del programa: {inicio}\n")
        archivo.write(f"Fin del programa: {fin}\n")
        archivo.write(f"Duración del programa: {duracion}\n")
        archivo.write(f"Cantidad de datos subidos: {contador_for}\n")
        archivo.write("Tablas subidas:\n")
        for dato in datos_for:
            archivo.write(f"-Server: {dato['server']} ")
            archivo.write(f"-Tabla: {dato['tabla']} ")
            archivo.write(f"-Estado: {dato['estado']}\n")
        archivo.write("\n")