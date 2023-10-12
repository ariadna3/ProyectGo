package staff

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/proyectoNovedades/servicios/constantes"
	"github.com/proyectoNovedades/servicios/models"
	"github.com/proyectoNovedades/servicios/proveedores"
	"github.com/proyectoNovedades/servicios/recursos"
	"github.com/proyectoNovedades/servicios/userGoogle"
	"github.com/proyectoNovedades/servicios/utils"
)

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
	userGoogle.ConnectMongoDb(client)
}

func InsertStaff(c *fiber.Ctx) error {
	fmt.Println("Ingreso de freelance")
	// validar el token
	error, codigo, email := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}
	err, usuario := userGoogle.GetInternUserITP(email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	// obtencion de datos
	staff := new(models.Staff)
	if err := c.BodyParser(staff); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	//Obtiene el ultimo Id
	coll := client.Database(constantes.Database).Collection(constantes.CollectionStaff)
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "idStaff", Value: -1}})
	cursor, _ := coll.Find(context.Background(), filter, opts)
	var results []models.Staff
	cursor.All(context.Background(), &results)
	var id int
	if len(results) == 0 {
		id = 0
	} else {
		id = results[0].IdStaff + 1
	}

	// Settea la nomina en S
	staff.Nomina = "S"

	// Verifica si existe el legajo o cuit en recursos y proveedores
	if staff.CUIT != "" {
		coll := client.Database(constantes.Database).Collection(constantes.CollectionProveedor)
		filter := bson.D{{Key: "cuit", Value: staff.CUIT}}
		cursor, _ := coll.Find(context.Background(), filter)
		var results []proveedores.Proveedores
		cursor.All(context.Background(), &results)
		if len(results) == 0 {
			return c.Status(fiber.StatusNotFound).SendString("El CUIT no existe en proveedores")
		}
		staff.Apellido = results[0].RazonSocial
		staff.Nombre = ""
	} else {
		coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
		filter := bson.D{{Key: "legajo", Value: staff.NroLegajo}}
		cursor, _ := coll.Find(context.Background(), filter)
		var results []recursos.Recursos
		cursor.All(context.Background(), &results)
		if len(results) == 0 {
			return c.Status(fiber.StatusNotFound).SendString("El legajo no existe en recursos")
		}
		staff.Apellido = results[0].Apellido
		staff.Nombre = results[0].Nombre
		staff.Gerente, err = strconv.Atoi(results[0].Gerente)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		staff.FechaIngreso = results[0].Fecha
		staff.SueldoFacturaMonto = float64(results[0].Sueldo)
	}

	// Ingresar Freelance
	staff.IdStaff = id
	result, err := coll.InsertOne(context.TODO(), staff)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	utils.SaveStaffInsert(usuario, *staff, "insert")
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.Status(200).JSON(staff)

}
