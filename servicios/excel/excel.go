package excel

import (
	"context"
	"fmt"
<<<<<<< HEAD
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
=======
	"log"
	"os"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/proyectoNovedades/servicios/constantes"
	"github.com/proyectoNovedades/servicios/novedades"
	"github.com/proyectoNovedades/servicios/recursos"
	"github.com/proyectoNovedades/servicios/userGoogle"
)

var client *mongo.Client
>>>>>>> a759c1746b01548e8f846a0b71b3b0fead1c1ac7

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
	userGoogle.ConnectMongoDb(client)
}

<<<<<<< HEAD
func GetExcelFile(c *fiber.Ctx) error {
	fmt.Println("GetExcelFile")

	//validar el token
=======
// Crear excel
func GetExcelFile(c *fiber.Ctx) error {
	fmt.Println("GetExcelFile")
	// validar el token
>>>>>>> a759c1746b01548e8f846a0b71b3b0fead1c1ac7
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)
<<<<<<< HEAD
	cursor, err := coll.Find(context.TODO(), bson.M{})
=======
	// {$and: [{descripcion:{$exists:1}}, {descripcion:{$ne:""}}, {usuario:{$exists:1}},{usuario:{$ne: ""}}]}
	usuarioExist := bson.D{{Key: "usuario", Value: bson.M{"$exists": 1}}}
	usuarioNotEmpty := bson.D{{Key: "usuario", Value: bson.M{"$ne": ""}}}
	descripcionExist := bson.D{{Key: "descripcion", Value: bson.M{"$exists": 1}}}
	descripcionNotEmpty := bson.D{{Key: "descripcion", Value: bson.M{"$ne": ""}}}

	filter := bson.M{"$and": bson.A{usuarioExist, usuarioNotEmpty, descripcionExist, descripcionNotEmpty}}
	opts := options.Find().SetSort(bson.D{{Key: "descripcion", Value: 1}, {Key: "usuario", Value: 1}})

	cursor, err := coll.Find(context.TODO(), filter, opts)
>>>>>>> a759c1746b01548e8f846a0b71b3b0fead1c1ac7
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var novedades []novedades.Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		return c.Status(503).SendString(err.Error())
	}

<<<<<<< HEAD
	err = DatosExcel(novedades)
=======
	err = datosExcel(novedades)
>>>>>>> a759c1746b01548e8f846a0b71b3b0fead1c1ac7
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error al crear archivo: " + err.Error())
	}

	return c.Status(fiber.StatusOK).SendFile(os.Getenv("EXCEL_FILE"))
<<<<<<< HEAD

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
=======
}

// ingresar datos a un excel
func datosExcel(novedadesArr []novedades.Novedades) error {

	// Abrir archivo de excel
	file, err := excelize.OpenFile(os.Getenv("EXCEL_FILE"))
	if err != nil {
		file = excelize.NewFile()
		file.SetSheetName("Sheet1", constantes.PestanaGeneral)
		file.NewSheet(constantes.PestanaHorasExtras)
		file.NewSheet(constantes.PestanaLicencias)
		initializeExcel(file)
	}
	var rowGeneral int = 3
	//var rowHorasExtras int = 3
	var rowLicencias int = 3

	for _, item := range novedadesArr {
		fmt.Print(item.IdSecuencial)
		fmt.Println(" " + item.Descripcion)
		if item.IdSecuencial == 298 {
			fmt.Print(item.IdSecuencial)
			fmt.Println(" " + item.Descripcion)
		}
		var pasosWorkflow novedades.PasosWorkflow
		coll := client.Database(constantes.Database).Collection(constantes.CollectionPasosWorkflow)
		err := coll.FindOne(context.TODO(), bson.M{"tipo": item.Descripcion}).Decode(&pasosWorkflow)
		if pasosWorkflow.TipoExcel == constantes.DescAnticipo {
			err = anticipoPrestamo(file, item, rowGeneral, 1)
			if err == nil {
				rowGeneral = rowGeneral + 1
			}
		}
		if pasosWorkflow.TipoExcel == constantes.DescSueldoNuevo {
			err = nuevoSueldo(file, item, rowGeneral)
			if err == nil {
				rowGeneral = rowGeneral + 1
			}
		}
		if pasosWorkflow.TipoExcel == constantes.DescLicencia {
			err = licencias(file, item, rowLicencias)
			if err == nil {
				rowLicencias = rowLicencias + 1
			}
		}
		if pasosWorkflow.TipoExcel == constantes.DescPrestamo {
			err = anticipoPrestamo(file, item, rowGeneral, 6)
			if err == nil {
				rowGeneral = rowGeneral + 1
			}
		}
	}

	// guardar archivo
	err = file.SaveAs(os.Getenv("EXCEL_FILE"))
	if err != nil {
		log.Printf("No se pudo guardar el archivo de Excel por el error %s", err.Error())
		return err
	}
>>>>>>> a759c1746b01548e8f846a0b71b3b0fead1c1ac7

	return nil
}

<<<<<<< HEAD
func nuevosSueldo(fieldValue reflect.Value, file excelize.File, rowSave int, indexValue int) error {
	file.SetCellValue("novedades", fmt.Sprintf("%s%d", excelize.ToAlphaString(indexValue), rowSave), fieldValue.Interface())
=======
func nuevoSueldo(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0)
	if err != nil {
		return err
	}

	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), novedad.ImporteTotal)
	if strings.Contains(novedad.Descripcion, "retroactivo") {
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", row), "SI")
	}

	return nil
}

func anticipoPrestamo(file *excelize.File, novedad novedades.Novedades, row int, cuotas int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0)
	if err != nil {
		return err
	}

	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("H%d", row), novedad.ImporteTotal)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("I%d", row), novedad.ImporteTotal/float64(cuotas))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("J%d", row), cuotas)

	return nil
}

func gimnasio(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0)
	if err != nil {
		return err
	}

	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("K%d", row), novedad.ImporteTotal)

	return nil
}

func idioma(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0)
	if err != nil {
		return err
	}

	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("L%d", row), novedad.ImporteTotal)

	return nil
}

func tarjetaBeneficio(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0)
	if err != nil {
		return err
	}

	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("M%d", row), novedad.ImporteTotal)

	return nil
}

func licencias(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0)
	if err != nil {
		return err
	}
	fechaDesdeDate, err := time.Parse(constantes.FormatoFecha, novedad.FechaDesde)
	if err != nil {
		return err
	}
	fechaHastaDate, err := time.Parse(constantes.FormatoFecha, novedad.FechaHasta)
	if err != nil {
		return err
	}
	diferenciaFechas := int(fechaHastaDate.Sub(fechaDesdeDate).Hours() / 24)

	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("A%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("B%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("C%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("D%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("E%d", row), diferenciaFechas)

	return nil
}

func initializeExcel(file *excelize.File) error {
	// General
	// Unir celdas de general para sueldos y prestamos
	file.MergeCell(constantes.PestanaGeneral, "E1", "G1")
	file.MergeCell(constantes.PestanaGeneral, "H1", "J1")

	// Ingresar los nombres de las celdas en general
	file.SetCellValue(constantes.PestanaGeneral, "A2", "TIPO DE NOVEDAD")
	file.SetCellValue(constantes.PestanaGeneral, "B2", "LEGAJO")
	file.SetCellValue(constantes.PestanaGeneral, "C2", "APELLIDO")
	file.SetCellValue(constantes.PestanaGeneral, "D2", "NOMBRE")
	file.SetCellValue(constantes.PestanaGeneral, "E2", "NUEVO SUELDO")
	file.SetCellValue(constantes.PestanaGeneral, "F2", "AJUSTE RETROACTIVIDAD")
	file.SetCellValue(constantes.PestanaGeneral, "G2", "DEVENGAMIENTO")
	file.SetCellValue(constantes.PestanaGeneral, "H2", "MONTO TOTAL")
	file.SetCellValue(constantes.PestanaGeneral, "I2", "MONTO CUOTA")
	file.SetCellValue(constantes.PestanaGeneral, "J2", "TOTAL CUOTAS")
	file.SetCellValue(constantes.PestanaGeneral, "K2", "MONTO GIMNASIO")
	file.SetCellValue(constantes.PestanaGeneral, "L2", "MONTO IDIOMA")
	file.SetCellValue(constantes.PestanaGeneral, "M2", "MONTO TARJETA")
	file.SetCellValue(constantes.PestanaGeneral, "E1", "NUEVOS SUELDOS")
	file.SetCellValue(constantes.PestanaGeneral, "H1", "ANTICIPO / PRESTAMO")
	file.SetCellValue(constantes.PestanaGeneral, "K1", "GYM")
	file.SetCellValue(constantes.PestanaGeneral, "L1", "IDIOMA")
	file.SetCellValue(constantes.PestanaGeneral, "M1", "TARJETA BENEFICIO")

	// Horas extras
	// Unir celdas de horas extras para sueldos y prestamos
	file.MergeCell(constantes.PestanaHorasExtras, "D1", "H1")

	// Ingresar los nombres de las celdas en horas extras
	file.SetCellValue(constantes.PestanaHorasExtras, "A2", "LEGAJO")
	file.SetCellValue(constantes.PestanaHorasExtras, "B2", "APELLIDO")
	file.SetCellValue(constantes.PestanaHorasExtras, "C2", "NOMBRE")
	file.SetCellValue(constantes.PestanaHorasExtras, "D2", "AL 50% EXENTAS (CONCEPTO 2212)")
	file.SetCellValue(constantes.PestanaHorasExtras, "E2", "AL 100% EXENTAS (CONCEPTO 2220)")
	file.SetCellValue(constantes.PestanaHorasExtras, "F2", "HORAS FERIADO")
	file.SetCellValue(constantes.PestanaHorasExtras, "G2", "NOCTURNAS AL 50% (CONCEPTO 2213)")
	file.SetCellValue(constantes.PestanaHorasExtras, "H2", "NOCTURNAS AL 100% (CONCEPTO 2221)")

	// Licencias
	// Unir celdas de licencias para sueldos y prestamos
	file.MergeCell(constantes.PestanaLicencias, "D1", "E1")

	// Ingresar los nombres de las celdas en licencias
	file.SetCellValue(constantes.PestanaLicencias, "A2", "LEGAJO")
	file.SetCellValue(constantes.PestanaLicencias, "B2", "APELLIDO")
	file.SetCellValue(constantes.PestanaLicencias, "C2", "NOMBRE")
	file.SetCellValue(constantes.PestanaLicencias, "D2", "TIPO")
	file.SetCellValue(constantes.PestanaLicencias, "E2", "DIAS")

>>>>>>> a759c1746b01548e8f846a0b71b3b0fead1c1ac7
	return nil
}
