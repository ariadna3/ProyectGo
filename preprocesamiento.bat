go env -w GOOS=linux
go build
go env -w GOOS=windows
del proyectoNovedades.sh
ren proyectoNovedades proyectoNovedades.sh