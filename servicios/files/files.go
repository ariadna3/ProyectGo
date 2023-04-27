package files

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type Files struct {
	Nombre string `bson:"nombre"`
}

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
}

func UploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("document")

	if err == nil {

		c.SaveFile(file, fmt.Sprintf("./%s", file.Filename))
		c.SaveFile(file, fmt.Sprintf("./archivosSubidos/%s", file.Filename))
		c.SaveFile(file, fmt.Sprintf("/tmp/uploads_relative/%s", file.Filename))
	}
	return (c.SendString("Subido"))
}
