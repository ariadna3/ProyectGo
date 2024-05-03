## PRIMEROS PASOS
### Crear el query
Primero se debe crear el archivo **"query.json"** que contiene 

- Los query para que utilizara el programa *"querys"*
- Las URLs a las que le pegara el programa *"urls"*
- El token para conectarse a la api *"token"*
- Los datos para conectarse a la base de datos SQL Server *"server", "database", "user", "password"*

Puede utilizarse el archivo **"queryExample.json"** como un ejemplo. A su vez dentro de los query se debera agregar:

- El path al que se subiran los datos *"Path"*
- El server al que debera pegarle *"Server"* (Debe tener el mismo nombre que aparezca en la lista de URLs)
- El nombre de la tabla *"Tabla"* (Sirve para identificarla en log, por lo que no es necesario que refleje un dato real)
- El query *"Query"* (Todos deben tener al final el **"FOR JSON AUTO"**)

### En caso de queres modificar la conecxion con la base de datos SQL Server

Actualmente si se llenan los parametros **"password"** y **"user"** del query el programa se logueara en la base de dato
Utilizando dichos datos. En caso contrario intentara hacerlo con el login de windows (Es decir si se dejan los parametros vacios)

### Al ejecutar el .exe
Para ejecutar el .exe se debera agregar como primer argumento la query de la que obtendra los datos. Por ejemplo

**./main.exe query.json**