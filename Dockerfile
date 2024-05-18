# Utiliza la imagen oficial de Golang 1.22
FROM golang:1.22

# Copia los archivos de tu proyecto a la imagen
COPY . /app

# Establece el directorio de trabajo
WORKDIR /app

# Instala las dependencias de tu proyecto (si las tienes)
RUN go mod download

# Expone el puerto en el que se ejecutará tu aplicación
EXPOSE 3000:3000

# Comando para ejecutar tu aplicación
CMD ["go", "run", "main.go"]