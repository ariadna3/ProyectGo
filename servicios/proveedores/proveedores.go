package proveedores

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

type Proveedores struct {
	IdProveedor int    `bson:"idProveedor"`
	NumeroDoc   string `bson:"numeroDoc"`
	RazonSocial string `bson:"razonSocial"`
}

var store *session.Store = session.New()
var dbUser *gorm.DB
var client *mongo.Client

var maxAge int32 = 86400 * 30 // 30 days
var isProd bool = false       // Set to true when serving over https

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
}

// ----Proveedores----
// insertar proveedor
func InsertProveedor(c *fiber.Ctx) error {
	proveedor := new(Proveedores)
	if err := c.BodyParser(proveedor); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	coll := client.Database("portalDeNovedades").Collection("proveedores")
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"idProveedor", -1}})

	cursor, _ := coll.Find(context.TODO(), filter, opts)

	var results []Proveedores
	cursor.All(context.TODO(), &results)

	proveedor.IdProveedor = results[0].IdProveedor + 1
	result, err := coll.InsertOne(context.TODO(), proveedor)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(proveedor)
}

// obtener proveedor por id
func GetProveedor(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("proveedores")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idProveedor": idNumber})
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var proveedor []Proveedores
	if err = cursor.All(context.Background(), &proveedor); err != nil {
		fmt.Print(err)
	}
	return c.JSON(proveedor)
}

// obtener todos los proveedores
func GetProveedorAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("proveedores")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Print(err)
	}
	var proveedor []Proveedores
	if err = cursor.All(context.Background(), &proveedor); err != nil {
		fmt.Print(err)
	}
	return c.JSON(proveedor)
}

// borrar proveedor por id
func DeleteProveedor(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("proveedores")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idProveedor": idNumber})
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.SendString("proveedor eliminado")
}
