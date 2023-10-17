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

func GetStaff(c *fiber.Ctx) error {
	fmt.Println("Obtener staff")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	//Obtener el id
	idStaff := c.Params("id")
	id, err := strconv.Atoi(idStaff)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	//Obtiene el staff
	var staff models.Staff
	coll := client.Database(constantes.Database).Collection(constantes.CollectionStaff)
	filter := bson.D{{Key: "idStaff", Value: id}}
	err = coll.FindOne(context.Background(), filter).Decode(&staff)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	return c.Status(200).JSON(staff)
}

func GetStaffList(c *fiber.Ctx) error {
	fmt.Println("Obtener lista de staff")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	//Obtiene todos los staff
	coll := client.Database(constantes.Database).Collection(constantes.CollectionStaff)
	filter := bson.D{}
	cursor, _ := coll.Find(context.Background(), filter)
	var results []models.Staff
	cursor.All(context.Background(), &results)
	if len(results) == 0 {
		return c.Status(fiber.StatusNotFound).SendString("No hay staff")
	}

	return c.Status(200).JSON(results)
}

func ReplaceStaff(c *fiber.Ctx) error {
	fmt.Println("Actualizar staff")
	// validar el token
	error, codigo, email := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}
	err, usuario := userGoogle.GetInternUserITP(email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	// obtencion de datos
	staffNew := new(models.Staff)
	if err := c.BodyParser(staffNew); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	// obtener el staff viejo
	staffOld := new(models.Staff)
	coll := client.Database(constantes.Database).Collection(constantes.CollectionStaff)
	filter := bson.D{{Key: "idStaff", Value: staffNew.IdStaff}}
	err = coll.FindOne(context.Background(), filter).Decode(&staffOld)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	// crear un mapa de diferencias
	staffOldMap := utils.StaffToMap(*staffOld)
	staffNewMap := utils.StaffToMap(*staffNew)
	diferencias := make(map[string]interface{})
	// comparar los mapas
	for key, value := range staffNewMap {
		if staffOldMap[key] != value {
			diferencias[key] = value
		}
	}
	// ingresa el campo id
	diferencias["idFreelance"] = staffNew.IdStaff

	// Actualizar staff
	update := bson.D{{Key: "$set", Value: staffNew}}
	_, err = coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	// Ingresar Historial
	err = utils.SaveMapInsertStaff(usuario, diferencias, "update")
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	return c.Status(200).JSON(staffNew)
}

func DeleteStaff(c *fiber.Ctx) error {
	fmt.Println("Eliminar staff")
	// validar el token
	error, codigo, email := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}
	err, usuario := userGoogle.GetInternUserITP(email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	// obtener id
	idStaff := c.Params("id")
	id, err := strconv.Atoi(idStaff)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	// obtener el staff viejo
	staffOld := new(models.Staff)
	coll := client.Database(constantes.Database).Collection(constantes.CollectionStaff)
	filter := bson.D{{Key: "idStaff", Value: id}}
	err = coll.FindOne(context.Background(), filter).Decode(&staffOld)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	// eliminar staff
	result, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).SendString("No se elimino ningun staff")
	}

	// Ingresar Historial
	utils.SaveStaffInsert(usuario, *staffOld, "delete")
	return c.Status(200).JSON(result)
}

func GetStaffHistorial(c *fiber.Ctx) error {
	fmt.Println("Obtener historial de staff")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	//Obtiene todos los staff
	results, err := utils.GetStaffHistorial()
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	return c.Status(200).JSON(results)
}

func GetStaffHistorialById(c *fiber.Ctx) error {
	fmt.Println("Obtener historial de staff por id")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	//Obtiene el id
	idStaff := c.Params("id")
	id, err := strconv.Atoi(idStaff)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	//Obtiene todos los staff
	results, err := utils.GetStaffHistorialById(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	return c.Status(200).JSON(results)
}
