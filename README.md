# Portal de novedades Backend

Esta es una pequeña descripción de lo que hay que hacer previo a ejecutar el programa

## Verificar archivos
### Variables de ambiente
Se deben crear el archivo .env que contenga las variables de ambiente. A continuacion se describe la utilidad de cada variable

- MONGOURI: Es la uri para conectarse a la base de datos mongoDB
- MYSQLURI: Es la uri para conectarse a la base de datos SQL
- GOOGLEKEY: Es la key para certificarse en google 
- GOOGLESEC: Es el secret para certificarse en google
- GOOGLECALLBACK: Es la pagina a la que le hace callback cuando obtiene los datos de google
- PUERTO: es el puerto en el que se expondra el backend este debe tener el siguiente formato ":3000"
- PUERTOCORS: Es el puerto desde el que se quiere conectar al backend en caso de querer conectarse desde otra aplicacion web
- USER_EMAIL: Email desde el que se enviaran Mails 
- USER_PASSWORD: Contraseña del usuario de los mails 
- USER_PORT: Puerto del POP para enviar los datos (recomendable "587")
- USER_HOST: Host del POP para enviar los datos (recomendable "smtp.gmail.com")
### carpeta de archivos
Debes crear la carpeta archivosSubidos donde se guardaran los datos
### Ejemplo de mail
Debes crear el archivo email.txt el cual usará el programa como plantilla para los mails que deba enviar avisando cuando se suba una novedad. Este debe tener el siguiente formato

"Asunto|

Mensaje"

El asunto y el mensaje se separan con un "|" y un salto de linea. A su vez se podra hacer uso de la descripcion, usuario, motivo y comentarios añadiendo al mensaje las siguientes palabras claves en el lugar deseado

- %D -> descripcion
- %S -> usuario
- %M -> motivo
- %C -> comentarios