package freelances

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/proyectoNovedades/servicios/constantes"
	"github.com/proyectoNovedades/servicios/userGoogle"
)

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
	userGoogle.ConnectMongoDb(client)
}

// ----Freelances----
type Freelances struct {
	IdFreelance     int       `bson:"idFreelance"`
	NroFreelance    int       `bson:"nroFreelance"`
	CUIT            string    `bson:"cuit"`
	Apellido        string    `bson:"apellido"`
	Nombre          string    `bson:"nombre"`
	FechaIngreso    time.Time `bson:"fechaIngreso"`
	Nomina          string    `bson:"nomina"`
	Gerente         int       `bson:"gerente"`
	Vertical        string    `bson:"vertical"`
	HorasMen        int       `bson:"horasMen"`
	Cargo           string    `bson:"cargo"`
	FacturaMonto    float64   `bson:"facturaMondo"`
	FacturaDesde    time.Time `bson:"facturaDesde"`
	FacturaADCuit   string    `bson:"facturaADCuit"`
	FacturaADMonto  float64   `bson:"facturaADMonto"`
	FacturaADDesde  time.Time `bson:"facturaADDesde"`
	B21Monto        float64   `bson:"b21Monto"`
	B21Desde        time.Time `bson:"b21Desde"`
	Comentario      string    `bson:"comentario"`
	Habilitado      string    `bson:"habilitado"`
	FechaBaja       time.Time `bson:"fechaBaja"`
	Cecos           []Rcc     `bson:"cecos"`
	Telefono        string    `bson:"telefono"`
	EmailLaboral    string    `bson:"emailLaboral"`
	EmailPersonal   string    `bson:"emailPersonal"`
	FechaNacimiento time.Time `bson:"fechaNacimiento"`
	Genero          string    `bson:"genero"`
	Nacionalidad    string    `bson:"nacionalidad"`
	DomCalle        string    `bson:"domCalle"`
	DomNumero       int       `bson:"domNumero"`
	DomPiso         int       `bson:"domPiso"`
	DomDepto        string    `bson:"domDepto"`
	DomLocalidad    string    `bson:"domLocalidad"`
	DomProvincia    string    `bson:"domProvincia"`
}

type Rcc struct {
	CcId         int     `bson:"ccId"`
	CcNum        int     `bson:"ccNum"`
	CcPorcentaje float32 `bson:"ccPorcentaje"`
	CcNombre     string  `bson:"ccNombre"`
	CcCliente    string  `bson:"ccCliente"`
}

func InsertFreelance(c *fiber.Ctx) error {
	fmt.Println("Ingreso de freelance")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	// obtencion de datos
	freelance := new(Freelances)
	if err := c.BodyParser(freelance); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	//Obtiene el ultimo Id
	coll := client.Database(constantes.Database).Collection(constantes.CollectionFreelance)
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "idFreelance", Value: -1}})
	cursor, _ := coll.Find(context.Background(), filter, opts)
	var results []Freelances
	cursor.All(context.Background(), &results)
	var idFreelance int
	if len(results) == 0 {
		idFreelance = 0
	} else {
		idFreelance = results[0].IdFreelance + 1
	}

	// Ingresar Freelance
	freelance.IdFreelance = idFreelance
	result, err := coll.InsertOne(context.TODO(), freelance)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.Status(200).JSON(freelance)
}

func GetFreelance(c *fiber.Ctx) error {
	fmt.Println("Get Freelance")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	idFreelance, _ := strconv.Atoi(c.Params("id"))
	coll := client.Database(constantes.Database).Collection(constantes.CollectionFreelance)
	filter := bson.D{{Key: "idFreelance", Value: idFreelance}}
	var result Freelances
	err := coll.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	return c.Status(200).JSON(result)
}

func GetFreelancesList(c *fiber.Ctx) error {
	fmt.Println("Get Freelances List")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionFreelance)
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "idFreelance", Value: 1}})
	cursor, _ := coll.Find(context.Background(), filter, opts)
	var results []Freelances
	err := cursor.All(context.Background(), &results)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	return c.Status(200).JSON(results)
}

func UpdateFreelance(c *fiber.Ctx) error {
	fmt.Println("Update Freelance")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	freelance := new(Freelances)
	if err := c.BodyParser(freelance); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionFreelance)
	filter := bson.D{{Key: "idFreelance", Value: freelance.IdFreelance}}
	update := bson.D{{Key: "$set", Value: freelance}}
	result, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	if result.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).SendString("No se encontro el freelance")
	}
	if result.ModifiedCount == 0 {
		return c.Status(fiber.StatusNotFound).SendString("No se modifico el freelance")
	}
	fmt.Printf("Matched %v documents and updated %v documents.\n", result.MatchedCount, result.ModifiedCount)
	return c.Status(200).JSON(freelance)
}

func DeleteFreelance(c *fiber.Ctx) error {
	fmt.Println("Delete Freelance")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	idFreelance, _ := strconv.Atoi(c.Params("id"))
	coll := client.Database(constantes.Database).Collection(constantes.CollectionFreelance)
	filter := bson.D{{Key: "idFreelance", Value: idFreelance}}
	result, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).SendString("No se encontro el freelance")
	}
	fmt.Printf("Deleted %v documents in the trainers collection\n", result.DeletedCount)
	return c.Status(200).JSON(idFreelance)
}
