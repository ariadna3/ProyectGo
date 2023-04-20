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
	Porc        []P       `bson:"p"`
}

type P struct {
	Cc     string  `bson:"cc"`
	PorcCC float32 `bson:"porcCC"`
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
	return c.JSON(fecha)
}

// ----Recursos----
// insertar recurso
func InsertRecurso(c *fiber.Ctx) error {
	recurso := new(Recursos)
	if err := c.BodyParser(recurso); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	recurso.Fecha, _ = time.Parse("02/01/2006", recurso.FechaString)
	fmt.Print(recurso.Fecha.Month())

	coll := client.Database("portalDeNovedades").Collection("recursos")
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"idRecurso", -1}})

	cursor, _ := coll.Find(context.TODO(), filter, opts)

	var results []Recursos
	cursor.All(context.TODO(), &results)

	recurso.IdRecurso = results[0].IdRecurso + 1
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
	cursor, err := coll.Find(context.TODO(), bson.M{"idRecurso": idNumber})
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
	}
	var recursos []Recursos
	if err = cursor.All(context.Background(), &recursos); err != nil {
		fmt.Print(err)
	}

	var recursosInterface []interface{}
	for _, recurso := range recursos {
		recursoI := map[string]interface{}{
			"IdRecurso": recurso.IdRecurso,
			"Nombre":    recurso.Nombre,
			"Apellido":  recurso.Apellido,
			"Legajo":    recurso.Legajo,
			"Mail":      recurso.Mail,
			"Fecha":     recurso.Fecha,
			"Sueldo":    recurso.Sueldo,
		}
		var sliceDePorcentajes []map[string]interface{}
		for _, p := range recurso.Porc {
			porcentaje := map[string]interface{}{
				"ccNum":     p.Cc,
				"ccNombre":  "No encontrado",
				"ccCliente": "No encontrado",
				"ccPorc":    p.PorcCC,
			}

			coll := client.Database("portalDeNovedades").Collection("centroDeCostos")
			codigoInt, _ := strconv.Atoi(p.Cc)
			filter := bson.D{{"codigo", codigoInt}}
			var result Cecos
			err = coll.FindOne(context.TODO(), filter).Decode(&result)
			if err != nil {
				fmt.Println(err)
			} else {
				porcentaje["ccNombre"] = result.Cliente
				porcentaje["ccCliente"] = result.Descripcion
			}
			sliceDePorcentajes = append(sliceDePorcentajes, porcentaje)
		}
		recursoI["Rcc"] = sliceDePorcentajes
		recursosInterface = append(recursosInterface, recursoI)
	}

	return c.JSON(recursosInterface)
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
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idRecurso": idNumber})
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.SendString("recurso eliminado")
}
