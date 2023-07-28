package excel

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gofiber/fiber/v2"

	"github.com/proyectoNovedades/servicios/constantes"
	"github.com/proyectoNovedades/servicios/novedades"
	"github.com/proyectoNovedades/servicios/userGoogle"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client
var novedad novedades.Novedades

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
	userGoogle.ConnectMongoDb(client)
}

func GetExcelFile(c *fiber.Ctx) error {
	fmt.Println("GetExcelFile")

	//validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var novedades []novedades.Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	err = DatosExcel(novedades)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error al crear archivo: " + err.Error())
	}

	return c.Status(fiber.StatusOK).SendFile(os.Getenv("EXCEL_FILE"))

}

// ingresar datos a un excel
func DatosExcel(novedad []novedades.Novedades) error {

	// abrir archivo de excel
	file, err := excelize.OpenFile(os.Getenv("EXCEL_FILE"))
	if err != nil {
		file = excelize.NewFile()

	}
	file.NewSheet(constantes.SheetGeneral)
	file.NewSheet(constantes.SheetHorasExtra)
	file.NewSheet(constantes.SheetLicencias)

	return nil
}

func nuevosSueldo(fieldValue reflect.Value, file excelize.File, rowSave int, indexValue int) error {
	file.SetCellValue("novedades", fmt.Sprintf("%s%d", excelize.ToAlphaString(indexValue), rowSave), fieldValue.Interface())
	return nil
}
