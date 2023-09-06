package novedades

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/proyectoNovedades/servicios/constantes"
	"github.com/proyectoNovedades/servicios/recursos"
	"github.com/proyectoNovedades/servicios/userGoogle"
)

const FinalDeLosPasos = "fin de los pasos"

type Novedades struct {
	AdjuntoMotivo         string           `bson:"adjuntoMotivo"`
	Adjuntos              []string         `bson:"adjuntos"`
	Cantidad              string           `bson:"cantidad"`
	CecosNuevo            string           `bson:"cecosNuevo"`
	Cliente               string           `bson:"cliente"`
	ClienteNuevo          bool             `bson:"clienteNuevo"`
	Comentarios           string           `bson:"comentarios"`
	ConceptoDeFacturacion string           `bson:"conceptoDeFacturacion"`
	Contacto              string           `bson:"contacto"`
	Departamento          string           `bson:"departamento"`
	Descripcion           string           `bson:"descripcion"`
	Distribuciones        []Distribuciones `bson:"distribuciones"`
	DistribucionesStr     string
	EnviarA               string              `bson:"enviarA"`
	Estado                string              `bson:"estado"`
	Fecha                 string              `bson:"fecha"`
	FechaDesde            string              `bson:"fechaDesde"`
	FechaHasta            string              `bson:"fechaHasta"`
	Freelance             bool                `bson:"freelance"`
	Hora                  string              `bson:"hora"`
	IdSecuencial          int                 `bson:"idSecuencial"`
	ImporteTotal          float64             `bson:"importeTotal"`
	Motivo                string              `bson:"motivo"`
	OrdenDeCompra         string              `bson:"ordenDeCompra"`
	Periodo               string              `bson:"periodo"`
	Plazo                 string              `bson:"plazo"`
	Prioridad             string              `bson:"prioridad"`
	Promovido             bool                `bson:"promovido"`
	Proveedor             string              `bson:"proveedor"`
	ProvNuevo             bool                `bson:"provNuevo"`
	Reclamo               bool                `bson:"reclamo"`
	Recursos              []RecursosNovedades `bson:"recursos"`
	RecursosStr           string
	Resumen               string     `bson:"resumen"`
	Tipo                  string     `bson:"tipo"`
	Usuario               string     `bson:"usuario"`
	RolUsuario            string     `bson:"rolUsuario"`
	Workflow              []Workflow `bson:"workflow"`
	Archivado             bool       `bson:"archivado"`
}

type TipoNovedad struct {
	IdSecuencial int    `bson:"idSecuencial"`
	Tipo         string `bson:"tipo"`
	Descripcion  string `bson:"descripcion"`
}

type Cecos struct {
	IdCecos     int    `bson:"idCecos"`
	Descripcion string `bson:"descripcioncecos"`
	Cliente     string `bson:"cliente"`
	Proyecto    string `bson:"proyecto"`
	Codigo      int    `bson:"codigo"`
	CuitCuil    int    `bson:"cuitcuil"`
}

type PackageOfCecos struct {
	Paquete []Cecos
}

type Distribuciones struct {
	Porcentaje float32 `bson:"porcentaje"`
	Cecos      Cecos   `bson:"cecos"`
}

type RecursosNovedades struct {
	Importe     int           `bson:"importe"`
	Comentarios string        `bson:"comentarios"`
	Recurso     string        `bson:"recurso"`
	Periodo     string        `bson:"periodo"`
	Porc        []P           `bson:"p"`
	SbActual    float64       `bson:"sbActual"`
	Retroactivo bool          `bson:"retroactivo"`
	HorasExtras []HorasExtras `bson:"horasExtras"`
}

type HorasExtras struct {
	Legajo   int    `bson:"legajo"`
	TipoDia  string `bson:"tipoDia"`
	TipoHora string `bson:"tipoHora"`
	Cantidad int    `bson:"cantidad"`
}

type P struct {
	Cc       string  `bson:"cc"`
	PorcCC   float32 `bson:"porcCC"`
	Cantidad int     `bson:"cantidad"`
}

type Workflow struct {
	Aprobador   string    `bson:"aprobador"`
	Estado      string    `bson:"estado"`
	Autorizador string    `bson:"autorizador"`
	Fecha       time.Time `bson:"fecha"`
	FechaStr    string
}

type PasosWorkflow struct {
	TipoExcel string `bson:"tipoExcel"`
	Tipo      string `bson:"tipo"`
	Pasos     []Paso `bson:"pasos"`
}

type Paso struct {
	Aprobador   string `bson:"aprobador"`
	Responsable string `bson:"responsable"`
}

type FileDto struct {
	Archivo string
}

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
	userGoogle.ConnectMongoDb(client)
}

// ----Novedades----
// insertar novedad
func InsertNovedad(c *fiber.Ctx) error {
	fmt.Println("InsertNovedad")
	// validar el token
	error, codigo, emailUser := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	//obtiene los datos
	novedad := new(Novedades)
	if err := c.BodyParser(novedad); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	if novedad.EnviarA != "" {
		enviarMail(*novedad)
	}

	// valida el estado
	if novedad.Estado != constantes.Pendiente && novedad.Estado != constantes.Aceptada && novedad.Estado != constantes.Rechazada {
		novedad.Estado = constantes.Pendiente
	}

	//le asigna un idSecuencial
	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)

	filter := bson.D{{}}
	opts := options.Find().SetSort(bson.D{{"idSecuencial", -1}})

	cursor, err := coll.Find(context.TODO(), filter, opts)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	var results []Novedades
	if err = cursor.All(context.TODO(), &results); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	if len(results) == 0 {
		novedad.IdSecuencial = 1
	} else {
		novedad.IdSecuencial = results[0].IdSecuencial + 1
	}

	//ingresa los archivos
	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println(err)
		fmt.Println("No se subieron archivos")
	} else {
		for _, fileHeaders := range form.File {
			for _, fileHeader := range fileHeaders {
				existeEnAdjuntos, _ := findInStringArray(novedad.Adjuntos, fileHeader.Filename)
				if !existeEnAdjuntos && novedad.AdjuntoMotivo != fileHeader.Filename {
					novedad.Adjuntos = append(novedad.Adjuntos, fileHeader.Filename)
				}
				idName := strconv.Itoa(novedad.IdSecuencial)
				c.SaveFile(fileHeader, os.Getenv("FOLDER_FILE")+"/"+idName+fileHeader.Filename)
			}
		}
	}

	//paso de strings Distribuciones y Recursos a Jsons
	if novedad.DistribucionesStr != "" {
		err = json.Unmarshal([]byte(novedad.DistribucionesStr), &novedad.Distribuciones)
		if err != nil {
			fmt.Println("Error al transformar Distribucion a Json")
			fmt.Println(novedad.DistribucionesStr)
		}
	}

	if novedad.RecursosStr != "" {
		err = json.Unmarshal([]byte(novedad.RecursosStr), &novedad.Recursos)
		if err != nil {
			fmt.Println("Error al transformar Recursos a Json")
			fmt.Println(novedad.RecursosStr)
		}
	}

	novedad.DistribucionesStr = ""
	novedad.RecursosStr = ""

	//valida los pasos del workflow
	err = validarPasos(novedad)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return c.Status(fiber.StatusNotFound).SendString("Tipo de novedad no encontrada")
		}
		fmt.Println(err)
	}

	//Ingresar rol del usuario
	err, usuario := userGoogle.GetInternUserITP(emailUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	novedad.RolUsuario = usuario.Rol
	novedad.Usuario = emailUser

	//inserta la novedad
	result, err := coll.InsertOne(context.TODO(), novedad)
	if err != nil {
		fmt.Print(err)
		fmt.Println("fail")
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.Status(200).JSON(novedad)
}

// obtener novedad por id
func GetNovedades(c *fiber.Ctx) error {
	fmt.Println("GetNovedades")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idSecuencial": idNumber})
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var novedades []Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	for index, element := range novedades {
		novedades[index].Resumen = resumenNovedad(element)
	}
	return c.Status(200).JSON(novedades)
}

// Busqueda con parametros Novedades
func GetNovedadFiltro(c *fiber.Ctx) error {
	fmt.Println("GetNovedadFiltro")
	// validar el token
	error, codigo, email := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)
	var busqueda bson.M = bson.M{}
	if c.Query("tipo") != "" {
		busqueda["tipo"] = bson.M{"$regex": c.Query("tipo"), "$options": "im"}
	}
	if c.Query("fecha") != "" {
		busqueda["fecha"] = bson.M{"$regex": c.Query("fecha"), "$options": "im"}
	}
	if c.Query("hora") != "" {
		busqueda["hora"] = bson.M{"$regex": c.Query("hora"), "$options": "im"}
	}
	if c.Query("usuario") != "" {
		busqueda["usuario"] = bson.M{"$regex": c.Query("usuario"), "$options": "im"}
	}
	if c.Query("proveedor") != "" {
		busqueda["proveedor"] = bson.M{"$regex": c.Query("proveedor"), "$options": "im"}
	}
	if c.Query("periodo") != "" {
		busqueda["periodo"] = bson.M{"$regex": c.Query("periodo"), "$options": "im"}
	}
	if c.Query("conceptoDeFacturacion") != "" {
		busqueda["conceptoDeFacturacion"] = bson.M{"$regex": c.Query("conceptoDeFacturacion"), "$options": "im"}
	}
	if c.Query("comentarios") != "" {
		busqueda["comentarios"] = bson.M{"$regex": c.Query("comentarios"), "$options": "im"}
	}
	if c.Query("cliente") != "" {
		busqueda["cliente"] = bson.M{"$regex": c.Query("cliente"), "$options": "im"}
	}
	if c.Query("estado") != "" {
		busqueda["estado"] = bson.M{"$regex": c.Query("estado"), "$options": "im"}
	}
	if c.Query("aprobador") != "" {
		busqueda["aprobador"] = bson.M{"$regex": c.Query("aprobador"), "$options": "im"}
	}
	if c.Query("departamento") != "" {
		busqueda["departamento"] = bson.M{"$regex": c.Query("departamento"), "$options": "im"}
	}
	if c.Query("rol") != "" {
		busqueda["rolUsuario"] = bson.M{"$regex": c.Query("rol"), "$options": "im"}
	}
	if c.Query("archivado") != "" {
		busqueda["archivado"] = c.QueryBool("true")
	}
	if c.Query("validos") != "" {
		gt100 := bson.D{{Key: "cecos.codigo", Value: bson.D{{Key: "$gt", Value: 100}}}}

		clienteValido := bson.D{{Key: "cliente", Value: bson.D{{Key: "$not", Value: bson.D{{Key: "$regex", Value: constantes.CecosNotValids}}}}}, {Key: "cliente", Value: bson.D{{Key: "$exists", Value: 1}, {Key: "$ne", Value: ""}}}}
		elematch := bson.D{{Key: "distribuciones", Value: bson.D{{Key: "$elemMatch", Value: gt100}}}}
		//{$or:[{cliente:{$not:{$regex: "(99999999999)"}}},{distribuciones:{$elemMatch:{"cecos.codigo":{$gt:100}}}}],cliente:{$exists:1}}
		busqueda["$or"] = bson.A{clienteValido, elematch}
	}
	if c.Query("estadoWF") != "" {
		coll2 := client.Database(constantes.Database).Collection(constantes.CollectionUserITP)
		var usuario userGoogle.UserITP
		err2 := coll2.FindOne(context.TODO(), bson.M{"email": email}).Decode(&usuario)
		if err2 != nil {
			return c.Status(fiber.StatusNotFound).SendString("Usuario no encontrada")
		}
		err, recurso := recursos.GetRecursoInterno(email, 0, 0)
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Gerente no encontrado")
		}
		if c.Query("estadoWF") == "all" {
			responsables := bson.A{}
			var orTodo bson.D = bson.D{}
			if c.Query("responsableWF") != "" {
				responsablesArray := strings.Split(c.Query("responsableWF"), "|")
				if contieneElArray(responsablesArray, "soporte") {
					responsables = append(responsables, bson.D{{Key: "aprobador", Value: recurso.Gerente}})
				}
				if contieneElArray(responsablesArray, "gerente") {
					responsables = append(responsables, bson.D{{Key: "aprobador", Value: strconv.Itoa(recurso.Legajo)}})
				}
				if contieneElArray(responsablesArray, "grupo") {
					responsables = append(responsables, bson.D{{Key: "aprobador", Value: usuario.Rol}})
				}
			} else {
				soporte := bson.D{{Key: "aprobador", Value: recurso.Gerente}}
				mail := bson.D{{Key: "aprobador", Value: strconv.Itoa(recurso.Legajo)}}
				grupo := bson.D{{Key: "aprobador", Value: usuario.Rol}}
				errorSop, _, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, os.Getenv("APROVE_ROL"))
				responsables = bson.A{mail, grupo}
				if errorSop == nil {
					responsables = bson.A{mail, grupo, soporte}
				}
			}

			orTodo = bson.D{{Key: "$or", Value: responsables}}

			busqueda["workflow"] = bson.D{{Key: "$elemMatch", Value: orTodo}}
		} else {
			estadoWFRegex := bson.M{"$regex": c.Query("estadoWF"), "$options": "im"}
			responsables := bson.A{}
			if c.Query("responsableWF") != "" {
				responsablesArray := strings.Split(c.Query("responsableWF"), "|")
				if contieneElArray(responsablesArray, "soporte") {
					responsables = append(responsables, bson.D{{Key: "$and", Value: bson.A{bson.D{{Key: "aprobador", Value: recurso.Gerente}}, bson.D{{Key: "estado", Value: estadoWFRegex}}}}})
				}
				if contieneElArray(responsablesArray, "gerente") {
					responsables = append(responsables, bson.D{{Key: "$and", Value: bson.A{bson.D{{Key: "aprobador", Value: strconv.Itoa(recurso.Legajo)}}, bson.D{{Key: "estado", Value: estadoWFRegex}}}}})
				}
				if contieneElArray(responsablesArray, "grupo") {
					responsables = append(responsables, bson.D{{Key: "$and", Value: bson.A{bson.D{{Key: "aprobador", Value: usuario.Rol}, {Key: "estado", Value: estadoWFRegex}}}}})
				}
			} else {
				andMail := bson.D{{Key: "$and", Value: bson.A{bson.D{{Key: "aprobador", Value: strconv.Itoa(recurso.Legajo)}}, bson.D{{Key: "estado", Value: estadoWFRegex}}}}}
				andGrupo := bson.D{{Key: "$and", Value: bson.A{bson.D{{Key: "aprobador", Value: usuario.Rol}, {Key: "estado", Value: estadoWFRegex}}}}}
				andSoporte := bson.D{{Key: "$and", Value: bson.A{bson.D{{Key: "aprobador", Value: recurso.Gerente}}, bson.D{{Key: "estado", Value: estadoWFRegex}}}}}
				errorSop, _, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, os.Getenv("APROVE_ROL"))
				responsables = bson.A{andMail, andGrupo}
				if errorSop == nil {
					responsables = bson.A{andMail, andGrupo, andSoporte}
				}
			}

			orTodo := bson.D{{Key: "$or", Value: responsables}}

			busqueda["workflow"] = bson.D{{Key: "$elemMatch", Value: orTodo}}
		}
	}
	fmt.Println(busqueda)
	cursor, err := coll.Find(context.TODO(), busqueda)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var novedades []Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	var nuevaListaNovedades []Novedades
	for _, element := range novedades {

		if c.Query("diferenciaFechas") != "" {
			if element.FechaDesde != "" && element.FechaDesde != "null" && len(element.FechaDesde) == 10 {
				dias, err := difDatesInDays(element)
				if err != nil {
					fmt.Println(element)
					fmt.Println(err)
				} else {
					diasMaximo, err := strconv.Atoi(c.Query("diferenciaFechas"))
					if err != nil {
						fmt.Println(element)
						fmt.Println("Error during conversion")
					} else {
						if dias <= diasMaximo {
							element.Resumen = resumenNovedad(element)
							nuevaListaNovedades = append(nuevaListaNovedades, element)
						}
					}

				}
			}
		} else {
			element.Resumen = resumenNovedad(element)
			nuevaListaNovedades = append(nuevaListaNovedades, element)
		}

	}
	return c.Status(200).JSON(nuevaListaNovedades)
}

// borrar novedad por id
func DeleteNovedad(c *fiber.Ctx) error {
	fmt.Println("DeleteNovedad")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idSecuencial": idNumber})
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	fmt.Printf("Deleted %v documents in the trainers collection\n", result.DeletedCount)
	return c.Status(200).SendString("novedad eliminada")
}

func UpdateEstadoNovedades(c *fiber.Ctx) error {
	fmt.Println("UpdateEstadoNovedades")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	//se obtiene el id
	idNumber, err := strconv.Atoi(c.Params("id"))
	fmt.Println(idNumber)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	//se obtiene el estado
	estado := c.Params("estado")
	novedad := new(Novedades)
	if err := c.BodyParser(novedad); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	//se conecta a la DB
	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)

	//Verifica que el estado sea uno valido
	if estado != constantes.Pendiente && estado != constantes.Aceptada && estado != constantes.Rechazada {
		return c.SendString("estado no valido")
	}

	//crea el filtro
	filter := bson.D{{Key: "idSecuencial", Value: idNumber}}

	//le dice que es lo que hay que modificar y con que
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "estado", Value: estado}}}}
	if novedad.Motivo != "" {
		update = bson.D{{Key: "$set", Value: bson.D{{Key: "estado", Value: estado}, {Key: "motivo", Value: novedad.Motivo}}}}
	}

	fmt.Println(update)
	fmt.Println(filter)

	//hace la modificacion
	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	//devuelve el resultado
	return c.Status(200).JSON(result)
}

func UpdateMotivoNovedades(c *fiber.Ctx) error {
	fmt.Println("UpdateMotivoNovedades")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	idNumber, _ := strconv.Atoi(c.Params("id"))
	novedad := new(Novedades)
	if err := c.BodyParser(novedad); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)

	filter := bson.D{{Key: "idSecuencial", Value: idNumber}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "motivo", Value: novedad.Motivo}}}}

	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	return c.Status(200).JSON(result)
}

func GetFiles(c *fiber.Ctx) error {
	fmt.Println("GetFiles")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	var novedad Novedades
	err := coll.FindOne(context.TODO(), bson.M{"idSecuencial": idNumber}).Decode(&novedad)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("novedad no encontrada")
	}

	if c.Query("nombre") != "" {
		nombre := c.Query("nombre")
		idName := strconv.Itoa(novedad.IdSecuencial)
		existeArchivo, _ := findInStringArray(novedad.Adjuntos, nombre)
		fmt.Println("No contiene (/): " + strconv.FormatBool(!strings.Contains(nombre, "/")))
		fmt.Println("No contiene mas de un punto: " + strconv.FormatBool(strings.Count(nombre, ".") == 1))
		fmt.Println("Existe en los adjuntos de la novedad: " + strconv.FormatBool(existeArchivo))
		fmt.Println("Existe en el adjunto motivo: " + strconv.FormatBool(novedad.AdjuntoMotivo == nombre))
		if !strings.Contains(nombre, "/") && strings.Count(nombre, ".") == 1 && (existeArchivo || novedad.AdjuntoMotivo == nombre) {
			return c.Status(200).SendFile(os.Getenv("FOLDER_FILE") + "/" + idName + nombre)
		} else {
			return c.Status(fiber.StatusBadRequest).SendString("nombre invalido")
		}
	}
	if c.Query("pos") != "" {
		posicion, _ := strconv.Atoi(c.Query("pos"))
		if len(novedad.Adjuntos) <= posicion {
			return c.Status(fiber.StatusBadRequest).SendString("posicion inexistente")
		}
		idName := strconv.Itoa(novedad.IdSecuencial)
		return c.Status(200).SendFile(os.Getenv("FOLDER_FILE") + "/" + idName + novedad.Adjuntos[posicion])
	} else {
		return c.Status(fiber.StatusBadRequest).SendString("debe especificar el archivo")
	}
	return c.Status(200).SendString("documento no encontrado")
}

// Agregar archivos a una notificacion
func UpdateFileAdd(c *fiber.Ctx) error {
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}
	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	var novedad Novedades
	err := coll.FindOne(context.TODO(), bson.M{"idSecuencial": idNumber}).Decode(&novedad)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("novedad no encontrada")
	}

	//ingresa los archivos
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusNoContent).SendString("archivo no encontrado")
	} else {
		fmt.Println(form.File)
		for _, fileHeaders := range form.File {
			for _, fileHeader := range fileHeaders {
				fmt.Println(fileHeader)
				existeEnAdjuntos, _ := findInStringArray(novedad.Adjuntos, fileHeader.Filename)
				if !existeEnAdjuntos && novedad.AdjuntoMotivo != fileHeader.Filename {
					novedad.Adjuntos = append(novedad.Adjuntos, fileHeader.Filename)
				}
				idName := strconv.Itoa(novedad.IdSecuencial)
				c.SaveFile(fileHeader, os.Getenv("FOLDER_FILE")+"/"+idName+fileHeader.Filename)
			}
		}
	}

	filter := bson.D{{Key: "idSecuencial", Value: idNumber}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "adjuntos", Value: novedad.Adjuntos}}}}

	_, err = coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	return c.Status(fiber.StatusOK).SendString("Subidos archivos correctamente")
}

// Eliminar archivos a una notificacion
func DeleteFile(c *fiber.Ctx) error {
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}
	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	var novedad Novedades
	err := coll.FindOne(context.TODO(), bson.M{"idSecuencial": idNumber}).Decode(&novedad)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("novedad no encontrada")
	}

	//ingresa los archivos
	var archivo FileDto
	if err := c.BodyParser(&archivo); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	var borrado bool = false
	for i, adjunto := range novedad.Adjuntos {
		if adjunto == archivo.Archivo {
			novedad.Adjuntos = append(novedad.Adjuntos[:i], novedad.Adjuntos[i+1:]...)
			direccionArchivo := os.Getenv("FOLDER_FILE") + "/" + (strconv.Itoa(novedad.IdSecuencial) + archivo.Archivo)
			os.Remove(direccionArchivo)
			borrado = true
		}
	}

	filter := bson.D{{Key: "idSecuencial", Value: idNumber}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "adjuntos", Value: novedad.Adjuntos}}}}

	_, err = coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	if borrado {
		return c.Status(fiber.StatusOK).SendString("Eliminado archivos correctamente")
	}
	return c.Status(fiber.StatusOK).SendString("El archivos no pudo eliminarse")
}

// ----Tipo Novedades----
func GetTipoNovedad(c *fiber.Ctx) error {
	fmt.Println("GetTipoNovedad")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database("portalDeNovedades").Collection("tipoNovedad")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var tipoNovedad []TipoNovedad

	if err = cursor.All(context.Background(), &tipoNovedad); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	return c.Status(200).JSON(tipoNovedad)
}

// ----Cecos----
// insertar cecos
func InsertCecos(c *fiber.Ctx) error {
	fmt.Println("InsertCecos")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	cecos := new(Cecos)
	if err := c.BodyParser(cecos); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	obtenerCuitCuil(cecos)

	coll := client.Database(constantes.Database).Collection(constantes.CollectionCecos)
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "idCecos", Value: -1}})

	cursor, err := coll.Find(context.TODO(), filter, opts)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	// obtiene el id del centro de costos
	var results []Cecos
	if err = cursor.All(context.TODO(), &results); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	if len(results) == 0 {
		cecos.IdCecos = 1
	} else {
		cecos.IdCecos = results[0].IdCecos + 1
	}

	// elimina el ceco si ya existe
	err = elCecoYaExiste(cecos.Descripcion)
	if err != nil {
		eliminarCeco(cecos.Descripcion)
	}

	// inserta el centro de costos nuevo
	result, err := coll.InsertOne(context.TODO(), cecos)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(cecos)
}

// insertar paquete de centros de costos
func InsertCecosPackage(c *fiber.Ctx) error {

	fmt.Println("Ingreso de paquete de centros de costos")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	//obtencion de datos
	packageCecos := new(PackageOfCecos)
	if err := c.BodyParser(packageCecos); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	fmt.Print("obtencion de datos ")
	fmt.Println(packageCecos)
	ingresarPaqueteDeCecos(*packageCecos)
	return c.Status(200).JSON(packageCecos)
}

// obtener todos los cecos
func GetCecosAll(c *fiber.Ctx) error {
	fmt.Println("GetCecosAll")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionCecos)
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var cecos []Cecos
	if err = cursor.All(context.Background(), &cecos); err != nil {
		c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	for index, _ := range cecos {
		obtenerCuitCuil(&cecos[index])
	}

	fmt.Println("procesado cecos")
	return c.JSON(cecos)
}

// obtener los cecos por codigo
func GetCecos(c *fiber.Ctx) error {
	fmt.Println("GetCecos")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionCecos)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"codigo": idNumber})
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var cecos []Cecos
	if err = cursor.All(context.Background(), &cecos); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	for index, _ := range cecos {
		obtenerCuitCuil(&cecos[index])
	}

	fmt.Println("procesado cecos")
	fmt.Println(cecos)
	return c.JSON(cecos)
}

func GetCecosFiltro(c *fiber.Ctx) error {
	fmt.Println("GetCecosFiltro")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionCecos)
	var busqueda bson.M = bson.M{}
	if c.Query("descripcion") != "" {
		busqueda["descripcion"] = bson.M{"$regex": c.Query("descripcion"), "$options": "im"}
	}
	if c.Query("cliente") != "" {
		busqueda["cliente"] = bson.M{"$regex": c.Query("cliente"), "$options": "im"}
	}
	if c.Query("proyecto") != "" {
		busqueda["proyecto"] = bson.M{"$regex": c.Query("proyecto"), "$options": "im"}
	}

	cursor, err := coll.Find(context.TODO(), busqueda)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var result []Cecos
	if err = cursor.All(context.Background(), &result); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	for index, _ := range result {
		obtenerCuitCuil(&result[index])
	}

	return c.JSON(result)
}

// ----Workflow----
// ingresar un workflow
func InsertWorkFlow(c *fiber.Ctx) error {
	fmt.Println("InsertWorkFlow")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	workflow := new(PasosWorkflow)
	if err := c.BodyParser(workflow); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	// inserta el centro de costos nuevo
	coll := client.Database(constantes.Database).Collection(constantes.CollectionPasosWorkflow)
	result, err := coll.InsertOne(context.TODO(), workflow)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(workflow)
}

// aprobar un workflow
func AprobarWorkflow(c *fiber.Ctx) error {
	fmt.Println("AprobarWorkflow")
	// validar el token
	error, codigo, email := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	var novedad Novedades
	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	err := coll.FindOne(context.TODO(), bson.M{"idSecuencial": idNumber}).Decode(&novedad)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	//Comprueba que la novedad no este aceptada o rechazada
	if novedad.Estado != constantes.Pendiente {
		return c.Status(fiber.ErrForbidden.Code).SendString("La novedad ya fue modificada")
	}

	//Busca el legajo del usuario registrado
	err, recurso := recursos.GetRecursoInterno(email, 0, 0)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Gerente no encontrado")
	}

	//comprueba que el usuario sea el autorizado
	aprobador := novedad.Workflow[len(novedad.Workflow)-1].Aprobador

	//1- comprobacion por grupo
	errGrupo, _, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, aprobador)

	//2- comprobacion por soporte
	comprobacionSop := false
	if os.Getenv("APROVE_ROL") != "" {
		errSop, _, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, os.Getenv("APROVE_ROL"))
		if errSop == nil {
			comprobacionSop = (aprobador == recurso.Gerente)
		}
	}

	if strconv.Itoa(recurso.Legajo) != aprobador && errGrupo != nil && !comprobacionSop {
		return c.Status(fiber.ErrForbidden.Code).SendString("Usuario no autorizado")
	}

	//Settea la aprobacion
	novedad.Workflow[len(novedad.Workflow)-1].Estado = constantes.Aceptada
	novedad.Workflow[len(novedad.Workflow)-1].Autorizador = email
	novedad.Workflow[len(novedad.Workflow)-1].Fecha = time.Now()
	novedad.Workflow[len(novedad.Workflow)-1].FechaStr = time.Now().Format(constantes.FormatoFecha)
	err = validarPasos(&novedad)
	if err != nil {
		if err.Error() == FinalDeLosPasos {
			novedad.Estado = constantes.Aceptada
			enviarMailWorkflow(novedad)
		}
	}
	//crea el filtro
	filter := bson.D{{Key: "idSecuencial", Value: idNumber}}

	//le dice que es lo que hay que modificar y con que
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "workflow", Value: novedad.Workflow}, {Key: "estado", Value: novedad.Estado}}}}

	//hace la modificacion
	_, err = coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(fiber.ErrBadRequest.Code).SendString(err.Error())
	}
	return c.JSON(novedad)
}

// rechazar un workflow
func RechazarWorkflow(c *fiber.Ctx) error {
	fmt.Println("RechazarWorkflow")
	// validar el token
	error, codigo, email := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	var novedad Novedades
	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	err := coll.FindOne(context.TODO(), bson.M{"idSecuencial": idNumber}).Decode(&novedad)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	//Comprueba que la novedad no este aceptada o rechazada
	if novedad.Estado != constantes.Pendiente {
		return c.Status(fiber.ErrForbidden.Code).SendString("La novedad ya fue modificada")
	}

	//Busca el legajo del usuario registrado
	err, recurso := recursos.GetRecursoInterno(email, 0, 0)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Gerente no encontrado")
	}

	//comprueba que el usuario sea el autorizado
	aprobador := novedad.Workflow[len(novedad.Workflow)-1].Aprobador
	//1- comprobacion por grupo
	errGrupo, _, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, aprobador)

	//2- comprobacion por soporte
	comprobacionSop := false
	if os.Getenv("APROVE_ROL") != "" {
		errSop, _, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, os.Getenv("APROVE_ROL"))
		if errSop == nil {
			comprobacionSop = (aprobador == recurso.Gerente)
		}
	}

	if strconv.Itoa(recurso.Legajo) != aprobador && errGrupo != nil && !comprobacionSop {
		return c.Status(fiber.ErrForbidden.Code).SendString("Usuario no autorizado")
	}

	//Settea el rechazo
	novedad.Workflow[len(novedad.Workflow)-1].Estado = constantes.Rechazada
	novedad.Workflow[len(novedad.Workflow)-1].Autorizador = email
	novedad.Workflow[len(novedad.Workflow)-1].Fecha = time.Now()
	novedad.Workflow[len(novedad.Workflow)-1].FechaStr = time.Now().Format(constantes.FormatoFecha)
	novedad.Estado = constantes.Rechazada
	enviarMailWorkflow(novedad)
	//crea el filtro
	filter := bson.D{{Key: "idSecuencial", Value: idNumber}}

	//le dice que es lo que hay que modificar y con que
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "workflow", Value: novedad.Workflow}, {Key: "estado", Value: novedad.Estado}}}}

	//hace la modificacion
	_, err = coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(fiber.ErrBadRequest.Code).SendString(err.Error())
	}
	return c.JSON(novedad)
}

func GetNovedadesPendientes(c *fiber.Ctx) error {
	fmt.Println("GetNovedadPendiente")
	// validar el token
	error, codigo, email := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionUserITP)
	var usuario userGoogle.UserITP
	err2 := coll.FindOne(context.TODO(), bson.M{"email": email}).Decode(&usuario)
	if err2 != nil {
		return c.Status(200).SendString("usuario no encontrada")
	}

	coll = client.Database(constantes.Database).Collection(constantes.CollectionNovedad)
	andMail := bson.D{{Key: "$and", Value: bson.A{bson.D{{Key: "aprobador", Value: email}}, bson.D{{Key: "estado", Value: constantes.Pendiente}}}}}
	andGrupo := bson.D{{Key: "$and", Value: bson.A{bson.D{{Key: "aprobador", Value: usuario.Rol}, {Key: "estado", Value: constantes.Pendiente}}}}}
	orTodo := bson.D{{Key: "$or", Value: bson.A{andMail, andGrupo}}}

	filter := bson.D{{Key: "workflow", Value: bson.D{{Key: "$elemMatch", Value: orTodo}}}}
	fmt.Println(filter)
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var novedades []Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	return c.JSON(novedades)
}

func resumenNovedad(novedad Novedades) string {
	var resumen string
	var resumenDict map[string]interface{}
	if novedad.Tipo == "PP" {
		resumenDict = map[string]interface{}{
			"Proveedor":      novedad.Proveedor,
			"Plazo":          novedad.Plazo,
			"ImporteTotal":   novedad.ImporteTotal,
			"Adjuntos":       novedad.Adjuntos,
			"Distribuciones": novedad.Distribuciones,
			"Comentarios":    novedad.Comentarios,
		}
	}
	if novedad.Tipo == "HE" || novedad.Tipo == "IG" {
		resumenDict = map[string]interface{}{
			"Cliente":      novedad.Cliente,
			"Periodo":      novedad.Periodo,
			"Descripcion":  novedad.Descripcion,
			"ImporteTotal": novedad.ImporteTotal,
			"Adjuntos":     novedad.Adjuntos,
			"Recursos":     novedad.Recursos,
			"Comentarios":  novedad.Comentarios,
		}
	}
	if novedad.Tipo == "FS" {
		resumenDict = map[string]interface{}{
			"Cliente":       novedad.Cliente,
			"Periodo":       novedad.Periodo,
			"Descripcion":   novedad.Descripcion,
			"ImporteTotal":  novedad.ImporteTotal,
			"Adjuntos":      novedad.Adjuntos,
			"Recursos":      novedad.Recursos,
			"OrdenDeCompra": novedad.OrdenDeCompra,
			"Comentarios":   novedad.Comentarios,
		}
	}
	if novedad.Tipo == "RH" {
		resumenDict = map[string]interface{}{
			"Descripcion":    novedad.Descripcion,
			"ImporteTotal":   novedad.ImporteTotal,
			"Adjuntos":       novedad.Adjuntos,
			"Recursos":       novedad.Recursos,
			"Distribuciones": novedad.Distribuciones,
			"Comentarios":    novedad.Comentarios,
		}
	}
	if novedad.Tipo == "NP" {
		resumenDict = map[string]interface{}{
			"Descripcion":  novedad.Descripcion,
			"Usuario":      novedad.Usuario,
			"Adjuntos":     novedad.Adjuntos,
			"Periodo":      novedad.Periodo,
			"Comentarios":  novedad.Comentarios,
			"Departamento": novedad.Departamento,
		}
	}
	resumenJson, err := json.Marshal(resumenDict)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		resumen = string(resumenJson)
	}
	return resumen
}

func difDatesInDays(novedad Novedades) (int, error) {
	fechaDesde, err := time.Parse("2006-01-02", novedad.FechaDesde)
	if err != nil {
		return 0, err
	}
	fechaHasta, err := time.Parse("2006-01-02", novedad.FechaHasta)
	if err != nil {
		return 0, err
	}
	fmt.Print(fechaHasta)
	fmt.Print(fechaDesde)
	fmt.Println(int(fechaHasta.Sub(fechaDesde).Hours() / 24))
	return int(fechaHasta.Sub(fechaDesde).Hours() / 24), nil
}

func findInStringArray(arrayString []string, palabra string) (bool, int) {
	for posicion, dato := range arrayString {
		if dato == palabra {
			return true, posicion
		}
	}
	return false, len(arrayString)
}

func enviarMail(novedad Novedades) {
	// Configuración de SMTP
	smtpHost := os.Getenv("USER_HOST")
	smtpPort := os.Getenv("USER_PORT")
	smtpUsername := os.Getenv("USER_EMAIL")
	smtpPassword := os.Getenv("USER_PASSWORD")
	emailFile := os.Getenv("USER_EMAIL_FILE")

	if emailFile != "" {
		datosComoBytes, err := ioutil.ReadFile(emailFile)
		if err != nil {
			log.Println(err.Error())
		}
		// convertir el arreglo a string
		datosComoString := string(datosComoBytes)
		// imprimir el string
		mailMessage := strings.Split(strings.Replace(datosComoString, "\n", "", 1), "|")
		mailMessage[1] = replaceStringWithData(mailMessage[1], novedad)

		// Mensaje de correo electrónico
		to := []string{novedad.EnviarA}
		from := os.Getenv("USER_EMAIL")
		toMsg := novedad.EnviarA
		subject := mailMessage[0]
		body := mailMessage[1]

		msg := ComposeMimeMail(toMsg, from, subject, body)

		// Autenticación y envío del correo electrónico
		auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUsername, to, msg)
		if err != nil {
			log.Println(err.Error())
		}
		log.Println("Correo electrónico enviado con éxito.")
	}
}

func enviarMailWorkflow(novedad Novedades) {
	// Configuración de SMTP
	smtpHost := os.Getenv("USER_HOST")
	smtpPort := os.Getenv("USER_PORT")
	smtpUsername := os.Getenv("USER_EMAIL")
	smtpPassword := os.Getenv("USER_PASSWORD")
	emailFile := os.Getenv("USER_EMAIL_FILE_WORKFLOW")

	if emailFile != "" {
		datosComoBytes, err := ioutil.ReadFile(emailFile)
		if err != nil {
			log.Println(err.Error())
		}
		// convertir el arreglo a string
		datosComoString := string(datosComoBytes)
		// imprimir el string
		mailMessage := strings.Split(strings.Replace(datosComoString, "\n", "", 1), "|")
		mailMessage[1] = replaceStringWithData(mailMessage[1], novedad)

		// Mensaje de correo electrónico
		to := []string{novedad.EnviarA}
		from := os.Getenv("USER_EMAIL")
		toMsg := novedad.EnviarA
		subject := mailMessage[0]
		body := mailMessage[1]

		msg := ComposeMimeMail(toMsg, from, subject, body)

		// Autenticación y envío del correo electrónico
		auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUsername, to, msg)
		if err != nil {
			log.Println(err.Error())
		}
		log.Println("Correo electrónico enviado con éxito.")
	}
}

func replaceStringWithData(message string, novedad Novedades) string {
	message = strings.ReplaceAll(message, "%D", novedad.Descripcion)
	message = strings.ReplaceAll(message, "%S", novedad.Usuario)
	message = strings.ReplaceAll(message, "%M", novedad.Motivo)
	message = strings.ReplaceAll(message, "%C", novedad.Comentarios)
	message = strings.ReplaceAll(message, "%E", novedad.Estado)
	return message
}

func formatEmailAddress(addr string) string {
	e, err := mail.ParseAddress(addr)
	if err != nil {
		return addr
	}
	return e.String()
}

func encodeRFC2047(str string) string {
	// use mail's rfc2047 to encode any string
	addr := mail.Address{Address: str}
	return strings.Trim(addr.String(), " <>")
}

func ComposeMimeMail(to string, from string, subject string, body string) []byte {
	header := make(map[string]string)
	header["From"] = formatEmailAddress(from)
	header["To"] = formatEmailAddress(to)
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	return []byte(message)
}

func validarPasos(novedad *Novedades) error {
	fmt.Print("Validando pasos, Actual: ")
	posicionActual := len(novedad.Workflow)
	fmt.Println(posicionActual)

	var pasosWorkflow PasosWorkflow
	coll := client.Database(constantes.Database).Collection(constantes.CollectionPasosWorkflow)
	err := coll.FindOne(context.TODO(), bson.M{"tipo": novedad.Descripcion}).Decode(&pasosWorkflow)
	fmt.Println("novedad tipo: " + novedad.Descripcion)
	fmt.Print("novedad encontrada: ")
	fmt.Println(pasosWorkflow)
	if err != nil {
		return err
	}
	listaDePasos := pasosWorkflow.Pasos
	if posicionActual == len(listaDePasos) {
		return errors.New(FinalDeLosPasos)
	}

	pasoActual := listaDePasos[posicionActual]
	var nuevoWorkflow Workflow
	if pasoActual.Aprobador == constantes.TipoGerente {
		err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
		if err != nil {
			return err
		}
		nuevoWorkflow.Aprobador = recurso.Gerente
	} else if pasoActual.Aprobador == constantes.TipoGrupo {
		nuevoWorkflow.Aprobador = pasoActual.Responsable
	} else {
		return errors.New("Paso invalido detectado")
	}
	nuevoWorkflow.Estado = constantes.Pendiente
	nuevoWorkflow.Autorizador = ""
	nuevoWorkflow.FechaStr = ""
	nuevoWorkflow.Fecha = time.Now()
	novedad.Workflow = append(novedad.Workflow, nuevoWorkflow)
	return nil
}

func obtenerCuitCuil(ceco *Cecos) {
	if ceco.CuitCuil == 0 {
		CuitCuilNuevo := strings.Split(ceco.Descripcion, "(")[len(strings.Split(ceco.Descripcion, "("))-1]
		CuitCuilNuevo = strings.Replace(CuitCuilNuevo, ")", "", 1)
		ceco.CuitCuil, _ = strconv.Atoi(CuitCuilNuevo)
	}
}

func elCecoYaExiste(descripcion string) error {
	coll := client.Database(constantes.Database).Collection(constantes.CollectionCecos)
	filter := bson.D{{Key: "descripcioncecos", Value: descripcion}}

	cursor, _ := coll.Find(context.TODO(), filter)

	var results []Cecos
	cursor.All(context.TODO(), &results)
	if len(results) != 0 {
		return errors.New("ya existe el centro de costos")
	}
	return nil
}

func eliminarCeco(descripcion string) error {
	coll := client.Database(constantes.Database).Collection(constantes.CollectionCecos)
	result, err := coll.DeleteOne(context.TODO(), bson.M{"descripcioncecos": descripcion})
	if err != nil {
		return err
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return nil
}

func ingresarPaqueteDeCecos(paqueteDeCecos PackageOfCecos) {
	fmt.Println("Eliminado de los cecos: ")
	for _, cecos := range paqueteDeCecos.Paquete {
		err := elCecoYaExiste(cecos.Descripcion)
		if err != nil {
			fmt.Print(cecos.Codigo)
			fmt.Print(", ")
			eliminarCeco(cecos.Descripcion)
		}
	}
	coll := client.Database(constantes.Database).Collection(constantes.CollectionCecos)

	//Obtiene el ultimo Id
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "idCecos", Value: -1}})

	cursor, _ := coll.Find(context.TODO(), filter, opts)

	var results []Cecos
	cursor.All(context.TODO(), &results)

	var ultimoId int

	if len(results) == 0 {
		ultimoId = 0
	} else {
		ultimoId = results[0].IdCecos + 1
	}

	//Empieza el setteo y subida
	arrayOfCecos := make([]interface{}, len(paqueteDeCecos.Paquete))
	for index, ceco := range paqueteDeCecos.Paquete {
		obtenerCuitCuil(&ceco)
		ceco.IdCecos = ultimoId
		ultimoId = ultimoId + 1
		arrayOfCecos[index] = ceco
		paqueteDeCecos.Paquete[index] = ceco
	}

	// Ingresa el recurso
	result, err := coll.InsertMany(context.TODO(), arrayOfCecos)
	if err != nil {
		// terminar ejecucion del recurso actual y avisar
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedIDs...)
}

func contieneElArray(arrayBuscador []string, buscado string) bool {
	for _, encontrado := range arrayBuscador {
		if encontrado == buscado {
			return true
		}
	}
	return false
}
