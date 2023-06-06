package proveedores

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

type Proveedores struct {
	IdProveedor int    `bson:"idProveedor"`
	Cuit        string `bson:"cuit"`
	RazonSocial string `bson:"razonSocial"`
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

// ----Proveedores----
// insertar proveedor
func InsertProveedor(c *fiber.Ctx) error {

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
	proveedor := new(Proveedores)
	if err := c.BodyParser(proveedor); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	// obtiene el ultimo id
	coll := client.Database("portalDeNovedades").Collection("proveedores")
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"idProveedor", -1}})

	cursor, _ := coll.Find(context.TODO(), filter, opts)

	var results []Proveedores
	cursor.All(context.TODO(), &results)

	if len(results) == 0 {
		proveedor.IdProveedor = 0
	} else {
		proveedor.IdProveedor = results[0].IdProveedor + 1
	}

	// inserta el proveedor
	result, err := coll.InsertOne(context.TODO(), proveedor)
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.Status(200).JSON(proveedor)
}

// obtener proveedor por id
func GetProveedor(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("proveedores")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idProveedor": idNumber})
	fmt.Println(coll)
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}
	var proveedor []Proveedores
	if err = cursor.All(context.Background(), &proveedor); err != nil {
		return c.Status(404).SendString(err.Error())
	}
	return c.Status(200).JSON(proveedor)
}

// obtener todos los proveedores
func GetProveedorAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("proveedores")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}
	var proveedor []Proveedores
	if err = cursor.All(context.Background(), &proveedor); err != nil {
		return c.Status(404).SendString(err.Error())
	}

	return c.Status(200).JSON(proveedor)
}

// borrar proveedor por id
func DeleteProveedor(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("proveedores")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idProveedor": idNumber})
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.Status(200).SendString("proveedor eliminado")
}
