package sustitutos

import (

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/proyectoNovedades/servicios/userGoogle"
)

type Sustitutos struct {
	IdSustituto int    `bson:"idustituto"`
	Gerente     string `bson:"gerente"`
	Legajo      int    `bson:"legajo"`
	FechaDesde  string `bson:"fechaDesde"`
	FechaHasta  string `bson:"fechaHasta"`
	Estado      string `bson:"estado"`
	Nombre      string `bson:"nombre"`
	Apellido    string `bson:"apellido"`
}

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
	userGoogle.ConnectMongoDb(client)
}
// ----Sustitutos----
// insertar sustituto


/*
//crea el filtro
filter := bson.D{{Key: "LegajoGerente", Value: strconv.Itoa(recursos.Legajo)},{Key: "estado", Value: true}}

//le dice que es lo que hay que modificar y con que
update := bson.D{{Key: "$set", Value: bson.D{{Key: "estado", Value: false}}}}

//hace la modificacion
_, err = coll.UpdateOne(context.TODO(), filter, update)
if err != nil {
	return c.Status(fiber.ErrBadRequest.Code).SendString(err.Error())
}
return c.JSON(novedad)*/
