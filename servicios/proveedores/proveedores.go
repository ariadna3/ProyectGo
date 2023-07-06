package proveedores

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"

	"github.com/proyectoNovedades/servicios/constantes"
	"github.com/proyectoNovedades/servicios/userGoogle"
)

const adminRequired = true
const adminNotRequired = false
const anyRol = ""

type Proveedores struct {
	IdProveedor int    `bson:"idProveedor"`
	NumeroDoc   string `bson:"cuit"`
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

	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), adminNotRequired, anyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	// obtencion de datos
	proveedor := new(Proveedores)
	if err := c.BodyParser(proveedor); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	// obtiene el ultimo id
	coll := client.Database(constantes.Database).Collection(constantes.CollectionProveedor)
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "idProveedor", Value: -1}})

	cursor, _ := coll.Find(context.TODO(), filter, opts)

	var results []Proveedores
	cursor.All(context.TODO(), &results)

	if len(results) == 0 {
		proveedor.IdProveedor = 0
	} else {
		proveedor.IdProveedor = results[0].IdProveedor + 1
	}

	// verifica la existencia del proveedor
	err := elProveedorYaExiste(proveedor.RazonSocial)
	if err != nil {
		eliminarProveedor(proveedor.RazonSocial)
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

	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), adminNotRequired, anyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionProveedor)
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

	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), adminNotRequired, anyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionProveedor)
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

	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), adminNotRequired, anyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionProveedor)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idProveedor": idNumber})
	if err != nil {
		return c.Status(404).SendString(err.Error())
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.Status(200).SendString("proveedor eliminado")
}

func elProveedorYaExiste(razonSocial string) error {
	coll := client.Database(constantes.Database).Collection(constantes.CollectionProveedor)
	filter := bson.D{{Key: "razonSocial", Value: razonSocial}}

	cursor, _ := coll.Find(context.TODO(), filter)

	var results []Proveedores
	cursor.All(context.TODO(), &results)
	if len(results) != 0 {
		return errors.New("ya existe el proveedor")
	}
	return nil
}

func eliminarProveedor(razonSocial string) error {
	coll := client.Database(constantes.Database).Collection(constantes.CollectionProveedor)
	result, err := coll.DeleteOne(context.TODO(), bson.M{"razonSocial": razonSocial})
	if err != nil {
		return err
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return nil
}
