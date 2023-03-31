package novedades

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Novedades struct {
	IdSecuencial          int                 `bson:"idSecuencial"`
	Tipo                  string              `bson:"tipo"`
	Fecha                 string              `bson:"fecha"`
	Hora                  string              `bson:"hora"`
	Usuario               string              `bson:"usuario"`
	Proveedor             string              `bson:"proveedor"`
	Periodo               string              `bson:"periodo"`
	ImporteTotal          float64             `bson:"importeTotal"`
	ConceptoDeFacturacion string              `bson:"conceptoDeFacturacion"`
	Adjuntos              []string            `bson:"adjuntos"`
	Distribuciones        []Distribuciones    `bson:"distribuciones"`
	Comentarios           string              `bson:"comentarios"`
	Promovido             bool                `bson:"promovido"`
	Cliente               string              `bson:"cliente"`
	Estado                string              `bson:"estado"`
	Motivo                string              `bson:"motivo"`
	EnviarA               string              `bson:"enviarA"`
	Contacto              string              `bson:"contacto"`
	Plazo                 string              `bson:"plazo"`
	Descripcion           string              `bson:"descripcion"`
	Recursos              []RecursosNovedades `bson:"recursos"`
	Cantidad              string              `bson:"cantidad"`
	FechaDesde            string              `bson:"fechaDesde"`
	FechaHasta            string              `bson:"fechaHasta"`
	OrdenDeCompra         string              `bson:"ordenDeCompra"`
	Resumen               string              `bson:"resumen"`
	Aprobador             string              `bson:"aprobador"`
}

const (
	Pendiente = "pendiente"
	Aceptada  = "aceptada"
	Rechazada = "rechazada"
)

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
}

type Distribuciones struct {
	Porcentaje float32 `bson:"porcentaje"`
	Cecos      Cecos   `bson:"cecos"`
}

type RecursosNovedades struct {
	Importe     int    `bson:"importe"`
	Comentarios string `bson:"comentarios"`
	Recurso     string `bson:"recurso"`
	Periodo     string `bson:"periodo"`
	Porc        []P    `bson:"p"`
}

type P struct {
	Cc       string  `bson:"cc"`
	PorcCC   float32 `bson:"porcCC"`
	Cantidad int     `bson:"cantidad"`
}

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
}

// ----Novedades----
// insertar novedad
func InsertNovedad(c *fiber.Ctx) error {
	novedad := new(Novedades)
	if err := c.BodyParser(novedad); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	if novedad.Estado != Pendiente && novedad.Estado != Aceptada && novedad.Estado != Rechazada {
		novedad.Estado = Pendiente
	}

	coll := client.Database("portalDeNovedades").Collection("novedades")

	filter := bson.D{{}}
	opts := options.Find().SetSort(bson.D{{"idSecuencial", -1}})

	cursor, error := coll.Find(context.TODO(), filter, opts)

	var results []Novedades
	cursor.All(context.TODO(), &results)

	fmt.Println(results)
	fmt.Println(client)
	fmt.Println(error)

	if len(results) == 0 {
		novedad.IdSecuencial = 1
	} else {
		novedad.IdSecuencial = results[0].IdSecuencial + 1
	}
	result, err := coll.InsertOne(context.TODO(), novedad)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(novedad)
}

// obtener novedad por id
func GetNovedades(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idSecuencial": idNumber})
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var novedades []Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		fmt.Print(err)
	}
	for index, element := range novedades {
		novedades[index].Resumen = resumenNovedad(element)
	}
	return c.JSON(novedades)
}

// Busqueda con parametros Novedades
func GetNovedadFiltro(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
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

	cursor, err := coll.Find(context.TODO(), busqueda)
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var novedades []Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		fmt.Print(err)
	}
	for index, element := range novedades {
		novedades[index].Resumen = resumenNovedad(element)
	}
	return c.JSON(novedades)
}

// obtener todas las novedades
func GetNovedadesAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Print(err)
	}
	var novedades []Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		fmt.Print(err)
	}
	for index, element := range novedades {
		novedades[index].Resumen = resumenNovedad(element)
	}
	return c.JSON(novedades)
}

// borrar novedad por id
func DeleteNovedad(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idSecuencial": idNumber})
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.SendString("novedad eliminada")
}

func UpdateEstadoNovedades(c *fiber.Ctx) error {
	//se obtiene el id
	idNumber, _ := strconv.Atoi(c.Params("idSecuencial"))
	//se obtiene el estado
	estado := c.Params("estado")
	novedad := new(Novedades)
	if err := c.BodyParser(novedad); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	//se conecta a la DB
	coll := client.Database("portalDeNovedades").Collection("novedades")

	//Verifica que el estado sea uno valido
	if estado != Pendiente && estado != Aceptada && estado != Rechazada {
		return c.SendString("estado no valido")
	}

	//crea el filtro
	filter := bson.D{{"idSecuencial", idNumber}}

	//le dice que es lo que hay que modificar y con que
	update := bson.D{{"$set", bson.D{{"estado", estado}}}}
	if novedad.Motivo != "" {
		update = bson.D{{"$set", bson.D{{"estado", estado}, {"motivo", novedad.Motivo}}}}
	}

	fmt.Println(estado)

	//hace la modificacion
	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		panic(err)
	}
	//devuelve el resultado
	return c.JSON(result)
}

func UpdateMotivoNovedades(c *fiber.Ctx) error {
	idNumber, _ := strconv.Atoi(c.Params("id"))
	novedad := new(Novedades)
	if err := c.BodyParser(novedad); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	coll := client.Database("portalDeNovedades").Collection("novedades")

	filter := bson.D{{"idSecuencial", idNumber}}
	update := bson.D{{"$set", bson.D{{"motivo", novedad.Motivo}}}}

	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		panic(err)
	}
	return c.JSON(result)
}

// ----Tipo Novedades----
func GetTipoNovedad(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("tipoNovedad")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	fmt.Println("tipos")
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var tipoNovedad []TipoNovedad
	fmt.Println(tipoNovedad)

	if err = cursor.All(context.Background(), &tipoNovedad); err != nil {
		fmt.Print(err)
	}
	return c.JSON(tipoNovedad)
}

// ----Cecos----
// insertar cecos
func InsertCecos(c *fiber.Ctx) error {
	cecos := new(Cecos)
	if err := c.BodyParser(cecos); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	coll := client.Database("portalDeNovedades").Collection("centroDeCostos")
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"idCecos", -1}})

	cursor, _ := coll.Find(context.TODO(), filter, opts)

	var results []Cecos
	cursor.All(context.TODO(), &results)

	if len(results) == 0 {
		cecos.IdCecos = 1
	} else {
		cecos.IdCecos = results[0].IdCecos + 1
	}
	result, err := coll.InsertOne(context.TODO(), cecos)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(cecos)
}

// obtener todos los cecos
func GetCecosAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("centroDeCostos")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Print(err)
	}
	var cecos []Cecos
	if err = cursor.All(context.Background(), &cecos); err != nil {
		fmt.Print(err)
	}
	return c.JSON(cecos)
}

// obtener los cecos por codigo
func GetCecos(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("centroDeCostos")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"codigo": idNumber})
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var cecos []Cecos
	if err = cursor.All(context.Background(), &cecos); err != nil {
		fmt.Print(err)
	}
	return c.JSON(cecos)
}

func GetCecosFiltro(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("centroDeCostos")
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
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var result []Cecos
	if err = cursor.All(context.Background(), &result); err != nil {
		fmt.Print(err)
	}
	return c.JSON(result)
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
		}
	}
	if novedad.Tipo == "RH" {
		resumenDict = map[string]interface{}{
			"Descripcion":  novedad.Descripcion,
			"ImporteTotal": novedad.ImporteTotal,
			"Adjuntos":     novedad.Adjuntos,
			"Recursos":     novedad.Recursos,
		}
	}
	if novedad.Tipo == "NP" {
		resumenDict = map[string]interface{}{
			"Descripcion": novedad.Descripcion,
			"Usuario":     novedad.Usuario,
			"Adjuntos":    novedad.Adjuntos,
			"Periodo":     novedad.Periodo,
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
