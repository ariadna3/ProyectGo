package actividades

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"

	"github.com/proyectoNovedades/servicios/userGoogle"
)

const adminRequired = true
const adminNotRequired = false
const anyRol = ""

type Actividades struct {
	IdActividad int    `bson:"idActividad"`
	Usuario     string `bson:"usuario"`
	Fecha       string `bson:"fecha"`
	Hora        string `bson:"hora"`
	Actividad   string `bson:"actividad"`
}

var store *session.Store = session.New()
var dbUser *gorm.DB
var client *mongo.Client

var maxAge int32 = 86400 * 30 // 30 days
var isProd bool = false       // Set to true when serving over https

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
	userGoogle.ConnectMongoDb(client)
}

// ----Actividades----
// insertar actividad
func InsertActividad(c *fiber.Ctx) error {

	//Obtencion de token
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		// El token no está presente
		return fiber.NewError(fiber.StatusUnauthorized, "No se proporcionó un token de autenticación")
	}

	// Parsea el token
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	err, codigo := userGoogle.ValidacionDeUsuarioPropio(adminNotRequired, anyRol, tokenString)
	if err != nil {
		if codigo != "" {
			codigoError, _ := strconv.Atoi(codigo)
			return c.Status(codigoError).SendString(err.Error())
		}
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	// obtencion de datos
	actividad := new(Actividades)
	if err := c.BodyParser(actividad); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	// obtiene el ultimo id
	coll := client.Database("portalDeNovedades").Collection("actividades")
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"idActividad", -1}})

	cursor, err := coll.Find(context.TODO(), filter, opts)
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}

	var results []Actividades
	if err = cursor.All(context.TODO(), &results); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	if len(results) == 0 {
		actividad.IdActividad = 0
	} else {
		actividad.IdActividad = results[0].IdActividad + 1
	}

	// inserta la actividad
	result, err := coll.InsertOne(context.TODO(), actividad)
	if err != nil {
		c.Status(404).SendString(err.Error())
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.Status(200).JSON(actividad)
}

// obtener actividad por id
func GetActividad(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("actividades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idActividad": idNumber})
	fmt.Println(coll)
	if err != nil {
		c.Status(404).SendString(err.Error())
	}
	var actividad []Actividades
	if err = cursor.All(context.Background(), &actividad); err != nil {
		c.Status(503).SendString(err.Error())
	}
	return c.Status(200).JSON(actividad)
}

// obtener todas las actividades
func GetActividadAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("actividades")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		c.Status(404).SendString(err.Error())
	}
	var actividad []Actividades
	if err = cursor.All(context.Background(), &actividad); err != nil {
		c.Status(503).SendString(err.Error())
	}
	return c.Status(200).JSON(actividad)
}

// borrar actividad por id
func DeleteActividad(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("actividades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idActividad": idNumber})
	if err != nil {
		c.Status(404).SendString(err.Error())
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.Status(200).SendString("actividad eliminada")
}
