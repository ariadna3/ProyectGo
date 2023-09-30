package utils

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/proyectoNovedades/servicios/constantes"
	"github.com/proyectoNovedades/servicios/models"
	"github.com/proyectoNovedades/servicios/userGoogle"
)

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
	userGoogle.ConnectMongoDb(client)
}

func SaveFreelanceInsert(usuario userGoogle.UserITP, freelance models.Freelances, tipo string) error {
	var historial models.HistorialFreelance
	historial.UsuarioEmail = usuario.Email
	historial.UsuarioNombre = usuario.Nombre
	historial.UsuarioApellido = usuario.Apellido
	historial.Freelance = freelance
	historial.Tipo = tipo
	historial.FechaHora = time.Now()

	coll := client.Database(constantes.Database).Collection(constantes.CollectionHistorial)
	// Obtener el ultimo id
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "idHistorial", Value: -1}})
	cursor, _ := coll.Find(context.Background(), filter, opts)
	var results []models.HistorialFreelance
	cursor.All(context.Background(), &results)
	if len(results) == 0 {
		historial.IdHistorial = 0
	} else {
		historial.IdHistorial = results[0].IdHistorial + 1
	}

	// Ingresar Historial
	result, err := coll.InsertOne(context.Background(), historial)
	if err != nil {
		return err
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return nil
}

func GetFreelanceHistorial() ([]models.HistorialFreelance, error) {
	coll := client.Database(constantes.Database).Collection(constantes.CollectionHistorial)
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "idHistorial", Value: -1}})
	cursor, err := coll.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	var results []models.HistorialFreelance
	cursor.All(context.Background(), &results)
	return results, nil
}

func GetFrelanceHistorialById(id int) (models.HistorialFreelance, error) {
	var historial models.HistorialFreelance
	coll := client.Database(constantes.Database).Collection(constantes.CollectionHistorial)
	filter := bson.D{{Key: "idHistorial", Value: id}}
	err := coll.FindOne(context.Background(), filter).Decode(&historial)
	if err != nil {
		return historial, err
	}
	return historial, nil
}
