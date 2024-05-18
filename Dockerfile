# Utiliza la imagen oficial de Golang 1.22
FROM golang:1.22

# Copia los de la pagina servicios a la carpeta /app de tu contenedor
COPY /servicios /app

# Copia el mod y sum a la carpeta /app de tu contenedor
COPY go.mod /app
COPY go.sum /app

# copia el archivo de configuracion a la carpeta /app de tu contenedor
COPY .env /app

# copia el main a la carpeta /app de tu contenedor
COPY main.go /app

# Establece el directorio de trabajo
WORKDIR /app

# Corre el tidy
RUN go mod tidy

# Instala las dependencias de tu proyecto (si las tienes)
RUN go mod download

# Expone el puerto en el que se ejecutará tu aplicación
EXPOSE 3000

# Compila tu aplicación
RUN go build -o main .

# Comando para ejecutar tu aplicación
CMD ["/app/main"]