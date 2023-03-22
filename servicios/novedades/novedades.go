package novedades

import (
	"context"
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
	IdCecos          int    `bson:"idCecos"`
	NCecos           string `bson:"nCecos"`
	DescripcionCecos string `bson:"descripcionCecos"`
	Cliente          string `bson:"cliente"`
}

type Distribuciones struct {
	Porcentaje float64 `bson:"porcentaje"`
	Cecos      Cecos   `bson:"cecos"`
}

type RecursosNovedades struct {
	Importe     int    `bson:"importe"`
	Comentarios string `bson:"comentarios"`
	Recurso     string `bson:"recurso"`
	Periodo     string `bson:"periodo"`
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

	novedad.IdSecuencial = results[0].IdSecuencial + 1
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
	return c.JSON(novedades)
}

// Busqueda con parametros Novedades
func GetNovedadFiltro(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
	var busqueda bson.M = bson.M{}
	if c.Query("tipo") != "" {
		busqueda["tipo"] = c.Query("tipo")
	}
	if c.Query("fecha") != "" {
		busqueda["fecha"] = c.Query("fecha")
	}
	if c.Query("hora") != "" {
		busqueda["hora"] = c.Query("hora")
	}
	if c.Query("usuario") != "" {
		busqueda["usuario"] = c.Query("usuario")
	}
	if c.Query("proveedor") != "" {
		busqueda["proveedor"] = c.Query("proveedor")
	}
	if c.Query("periodo") != "" {
		busqueda["periodo"] = c.Query("periodo")
	}
	if c.Query("conceptoDeFacturacion") != "" {
		busqueda["conceptoDeFacturacion"] = c.Query("conceptoDeFacturacion")
	}
	if c.Query("comentarios") != "" {
		busqueda["comentarios"] = c.Query("comentarios")
	}
	if c.Query("cliente") != "" {
		busqueda["cliente"] = c.Query("cliente")
	}
	if c.Query("estado") != "" {
		busqueda["estado"] = c.Query("estado")
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

	cecos.IdCecos = results[0].IdCecos + 1
	result, err := coll.InsertOne(context.TODO(), cecos)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(cecos)
}

// obtener todos los cecos
func GetCecosAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
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

// obtener los cecos por id
func GetCecos(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idSecuencial": idNumber})
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
