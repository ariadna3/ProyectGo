package recursos

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
}

type Recursos struct {
	idSecuencial int    `bson:"idSecuencial"`
	Usuario      string `bson:"usuario"`
	Legajo       int    `bson:"legajo"`
}

// ----Recursos----
// insertar recurso
func InsertRecurso(c *fiber.Ctx) error {
	recurso := new(Recursos)
	if err := c.BodyParser(recurso); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	coll := client.Database("portalDeNovedades").Collection("recursos")
	result, err := coll.InsertOne(context.TODO(), recurso)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(recurso)
}

// obtener recurso por id
func GetRecurso(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("recursos")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idSecuencial": idNumber})
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var recurso []Recursos
	if err = cursor.All(context.Background(), &recurso); err != nil {
		fmt.Print(err)
	}
	return c.JSON(recurso)
}

// obtener todos los recursos
func GetRecursoAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("recursos")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Print(err)
	}
	var recurso []Recursos
	if err = cursor.All(context.Background(), &recurso); err != nil {
		fmt.Print(err)
	}
	return c.JSON(recurso)
}

// borrar recurso por id
func DeleteRecurso(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("recursos")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idSecuencial": idNumber})
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.SendString("recurso eliminada")
}
