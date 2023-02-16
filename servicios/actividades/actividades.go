package actividades

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

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
}

// ----Actividades----
// insertar actividad
func InsertActividad(c *fiber.Ctx) error {
	actividad := new(Actividades)
	if err := c.BodyParser(actividad); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	coll := client.Database("portalDeNovedades").Collection("actividades")
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"idActividad", -1}})

	cursor, _ := coll.Find(context.TODO(), filter, opts)

	var results []Actividades
	cursor.All(context.TODO(), &results)

	actividad.IdActividad = results[0].IdActividad + 1
	result, err := coll.InsertOne(context.TODO(), actividad)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(actividad)
}

// obtener actividad por id
func GetActividad(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("actividades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idActividad": idNumber})
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var actividad []Actividades
	if err = cursor.All(context.Background(), &actividad); err != nil {
		fmt.Print(err)
	}
	return c.JSON(actividad)
}

// obtener todas las actividades
func GetActividadAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("actividades")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Print(err)
	}
	var actividad []Actividades
	if err = cursor.All(context.Background(), &actividad); err != nil {
		fmt.Print(err)
	}
	return c.JSON(actividad)
}

// borrar actividad por id
func DeleteActividad(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("actividades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idActividad": idNumber})
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.SendString("actividad eliminada")
}
