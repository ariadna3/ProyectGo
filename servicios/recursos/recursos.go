package recursos

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
}

type Recursos struct {
	IdRecurso   int       `bson:"idRecurso"`
	Nombre      string    `bson:"nombre"`
	Apellido    string    `bson:"apellido"`
	Legajo      int       `bson:"legajo"`
	Mail        string    `bson:"mail"`
	Fecha       time.Time `bson:"date"`
	FechaString string    `bson:"fechaString"`
	Sueldo      int       `bson:"sueldo"`
	Rcc         []P       `bson:"p"`
}

type P struct {
	CcNum     string  `bson:"cc"`
	CcPorc    float32 `bson:"porcCC"`
	CcNombre  string  `bson:"ccNombre"`
	CcCliente string  `bson:"ccCliente"`
}

type Cecos struct {
	IdCecos     int    `bson:"idCecos"`
	Descripcion string `bson:"descripcioncecos"`
	Cliente     string `bson:"cliente"`
	Proyecto    string `bson:"proyecto"`
	Codigo      int    `bson:"codigo"`
}

// fecha de ingreso
func GetFecha(c *fiber.Ctx) error {
	var fecha []string
	currentTime := time.Now()
	for i := 12; i >= 0; i-- {
		anio := strconv.Itoa(currentTime.Year())
		mes := strconv.Itoa(int(currentTime.Month()))
		fecha = append(fecha, mes+"-"+anio)
		currentTime = currentTime.AddDate(0, -1, 0)
	}
	return c.Status(200).JSON(fecha)
}

// ----Recursos----
// insertar recurso
func InsertRecurso(c *fiber.Ctx) error {
	recurso := new(Recursos)
	if err := c.BodyParser(recurso); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	//setea la fecha
	recurso.Fecha, _ = time.Parse("02/01/2006", recurso.FechaString)

	//Obtiene el ultimo Id
	coll := client.Database("portalDeNovedades").Collection("recursos")
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"idRecurso", -1}})

	cursor, _ := coll.Find(context.TODO(), filter, opts)

	var results []Recursos
	cursor.All(context.TODO(), &results)

	if len(results) == 0 {
		recurso.IdRecurso = 0
	} else {
		recurso.IdRecurso = results[0].IdRecurso + 1
	}

	//Obtiene los datos del ceco
	collCeco := client.Database("portalDeNovedades").Collection("centroDeCostos")

	for pos, ceco := range recurso.Rcc {
		intVar, err := strconv.Atoi(ceco.CcNum)
		if err != nil {
			fmt.Println(err)
			return c.Status(418).SendString(err.Error())
		}
		filter := bson.D{{"codigo", intVar}}

		var cecoEncontrado Cecos
		collCeco.FindOne(context.TODO(), filter).Decode(&cecoEncontrado)

		fmt.Print("Ceco encontrado: ")
		fmt.Println(cecoEncontrado)

		recurso.Rcc[pos].CcNombre = cecoEncontrado.Cliente
		recurso.Rcc[pos].CcCliente = cecoEncontrado.Descripcion
	}

	//Ingresa el recurso
	result, err := coll.InsertOne(context.TODO(), recurso)
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.Status(200).JSON(recurso)
}

// obtener recurso por id
func GetRecurso(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("recursos")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	var recurso Recursos
	err := coll.FindOne(context.TODO(), bson.D{{"idRecurso", idNumber}}).Decode(&recurso)
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
		return c.Status(404).SendString("No encontrado")
	}

	return c.Status(200).JSON(recurso)
}

// obtener todos los recursos
func GetRecursoAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("recursos")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}
	var recursos []Recursos
	if err = cursor.All(context.Background(), &recursos); err != nil {
		return c.Status(404).SendString(err.Error())
	}

	return c.Status(200).JSON(recursos)
}

// borrar recurso por id
func DeleteRecurso(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("recursos")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idRecurso": idNumber})
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.Status(200).SendString("recurso eliminado")
}
