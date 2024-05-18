# Utiliza la imagen oficial de Golang 1.22
FROM golang:1.22

# Copia los archivos de tu proyecto a la imagen
COPY . /app

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