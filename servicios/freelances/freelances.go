package freelances

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
	"github.com/proyectoNovedades/servicios/userGoogle"
	"github.com/proyectoNovedades/servicios/utils"
)

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
	userGoogle.ConnectMongoDb(client)
}

func InsertFreelance(c *fiber.Ctx) error {
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
	freelance := new(models.Freelances)
	if err := c.BodyParser(freelance); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	//Obtiene el ultimo Id
	coll := client.Database(constantes.Database).Collection(constantes.CollectionFreelance)
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "idFreelance", Value: -1}})
	cursor, _ := coll.Find(context.Background(), filter, opts)
	var results []models.Freelances
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

	utils.SaveFreelanceInsert(usuario, *freelance, "insert")
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
	var result models.Freelances
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
	var results []models.Freelances
	err := cursor.All(context.Background(), &results)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	return c.Status(200).JSON(results)
}

func UpdateFreelance(c *fiber.Ctx) error {
	fmt.Println("Update Freelance")
	// validar el token
	error, codigo, email := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}
	err, usuario := userGoogle.GetInternUserITP(email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	freelance := new(models.Freelances)
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
	err = utils.SaveFreelanceInsert(usuario, *freelance, "update")
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	fmt.Printf("Matched %v documents and updated %v documents.\n", result.MatchedCount, result.ModifiedCount)
	return c.Status(200).JSON(freelance)
}

func DeleteFreelance(c *fiber.Ctx) error {
	fmt.Println("Delete Freelance")
	// validar el token
	error, codigo, email := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}
	err, usuario := userGoogle.GetInternUserITP(email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	var freelanceBorrado models.Freelances

	idFreelance, _ := strconv.Atoi(c.Params("id"))
	coll := client.Database(constantes.Database).Collection(constantes.CollectionFreelance)
	filter := bson.D{{Key: "idFreelance", Value: idFreelance}}

	err = coll.FindOne(context.Background(), filter).Decode(&freelanceBorrado)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	result, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).SendString("No se encontro el freelance")
	}
	err = utils.SaveFreelanceInsert(usuario, freelanceBorrado, "delete")
	fmt.Printf("Deleted %v documents in the trainers collection\n", result.DeletedCount)
	return c.Status(200).JSON(idFreelance)
}

func GetFreelanceHistorial(c *fiber.Ctx) error {
	fmt.Println("Get Freelance Historial")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	results, err := utils.GetFreelanceHistorial()
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	return c.Status(200).JSON(results)
}

func GetFreelanceHistorialById(c *fiber.Ctx) error {
	fmt.Println("Get Freelance Historial By Id")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	idHistorial, _ := strconv.Atoi(c.Params("id"))
	result, err := utils.GetFrelanceHistorialById(idHistorial)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	return c.Status(200).JSON(result)
}
