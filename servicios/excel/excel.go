package excel

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gofiber/fiber/v2"

	"github.com/proyectoNovedades/servicios/constantes"
	"github.com/proyectoNovedades/servicios/novedades"
	"github.com/proyectoNovedades/servicios/recursos"
	"github.com/proyectoNovedades/servicios/userGoogle"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var novedad novedades.Novedades

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
	userGoogle.ConnectMongoDb(client)
}

// Crear excel
func GetExcelFile(c *fiber.Ctx) error {
	fmt.Println("GetExcelFile")

	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.PeopleOperation)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)

	cursor, err := coll.Find(context.TODO(), bson.M{})

	// {$and: [{descripcion:{$exists:1}}, {descripcion:{$ne:""}}, {usuario:{$exists:1}},{usuario:{$ne: ""}}]}
	usuarioExist := bson.D{{Key: "usuario", Value: bson.M{"$exists": 1}}}
	usuarioNotEmpty := bson.D{{Key: "usuario", Value: bson.M{"$ne": ""}}}
	descripcionExist := bson.D{{Key: "descripcion", Value: bson.M{"$exists": 1}}}
	descripcionNotEmpty := bson.D{{Key: "descripcion", Value: bson.M{"$ne": ""}}}
	EstadoNoRechazado := bson.D{{Key: "estado", Value: bson.M{"$ne": constantes.Rechazada}}}

	filter := bson.M{"$and": bson.A{usuarioExist, usuarioNotEmpty, descripcionExist, descripcionNotEmpty, EstadoNoRechazado}}
	opts := options.Find().SetSort(bson.D{{Key: "descripcion", Value: 1}, {Key: "usuario", Value: 1}})

	cursor, err = coll.Find(context.TODO(), filter, opts)

	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var novedades []novedades.Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	err = datosExcel(novedades, c.Query("fechaDesde"), c.Query("fechaHasta"))

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error al crear archivo: " + err.Error())
	}
	return c.SendFile(os.Getenv("EXCEL_FILE"))
}

// ingresar datos a un excel
func datosExcel(novedadesArr []novedades.Novedades, fechaDesde string, fechaHasta string) error {

	// Abrir archivo de excel
	os.Remove("EXCEL_FILE")
	file := excelize.NewFile()
	file.SetSheetName("Sheet1", constantes.PestanaGeneral)
	file.NewSheet(constantes.PestanaHorasExtras)
	file.NewSheet(constantes.PestanaLicencias)
	file.NewSheet(constantes.PestanaNovedades)
	initializeExcel(file)
	var rowGeneral int = 3
	var rowHorasExtras int = 3
	var rowLicencias int = 3
	var rowNovedades int = 3

	for _, item := range novedadesArr {
		if !verificacionNovedad(item, fechaDesde, fechaHasta) {
			continue
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
			err = nuevoSueldo(file, item, &rowGeneral)
			if err == nil {
				rowGeneral = rowGeneral + 1
			}
		}
		if pasosWorkflow.TipoExcel == constantes.DescSueldoNuevoMasivo {
			err = nuevoSueldoMasivo(file, item, &rowGeneral)
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
		if pasosWorkflow.TipoExcel == constantes.DescHorasExtras {
			horasExtras(file, item, &rowHorasExtras)
		}
		if pasosWorkflow.TipoExcel == constantes.DescPagos {
			pagos(file, item, &rowGeneral)
		}
		err = allNovedades(file, item, rowNovedades)
		if err == nil {
			rowNovedades = rowNovedades + 1
		}

	}

	// guardar archivo
	err := file.SaveAs(os.Getenv("EXCEL_FILE"))
	if err != nil {
		log.Printf("No se pudo guardar el archivo de Excel por el error %s", err.Error())
		return err
	}
	return nil
}

func nuevoSueldo(file *excelize.File, novedad novedades.Novedades, row *int) error {
	for _, recursoInterno := range novedad.Recursos {
		datosUsuario := strings.Split(recursoInterno.Recurso, "(")
		if len(datosUsuario) == 2 {
			datosUsuario[1] = strings.ReplaceAll(datosUsuario[1], ")", "")
			legajo, err := strconv.Atoi(datosUsuario[1])
			if err != nil {
				return err
			}
			err, recurso := recursos.GetRecursoInterno("", 0, legajo)
			if err != nil {
				return err
			}
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", *row), novedad.IdSecuencial)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", *row), novedad.Descripcion)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", *row), recurso.Legajo)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", *row), recurso.Nombre)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", *row), recurso.Apellido)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", *row), recursoInterno.Importe)
			if strings.Contains(novedad.Descripcion, "retroactivo") {
				file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("G%d", *row), "SI")
			}
			*row = *row + 1
		}

	}

	*row = *row - 1

	return nil
}

func nuevoSueldoMasivo(file *excelize.File, novedad novedades.Novedades, row *int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}

	for _, distribucion := range novedad.Distribuciones {
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", *row), novedad.IdSecuencial)
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", *row), novedad.Descripcion)
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", *row), recurso.Legajo)
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", *row), distribucion.Cecos.Cliente)
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", *row), "")
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", *row), fmt.Sprintf("%v", distribucion.Porcentaje)+"%")
		*row = *row + 1
	}
	*row = *row - 1
	return nil
}

func anticipoPrestamo(file *excelize.File, novedad novedades.Novedades, row int, cuotas int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}

	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("I%d", row), novedad.ImporteTotal)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("J%d", row), novedad.ImporteTotal/float64(cuotas))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("K%d", row), cuotas)

	return nil
}

func gimnasio(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}

	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("L%d", row), novedad.ImporteTotal)

	return nil
}

func idioma(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}

	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("M%d", row), novedad.ImporteTotal)

	return nil
}

func tarjetaBeneficio(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}

	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("N%d", row), novedad.ImporteTotal)

	return nil
}

func licencias(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}
	diferenciaFechas, _ := strconv.Atoi(novedad.Cantidad)

	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("A%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("B%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("C%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("D%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("E%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("F%d", row), diferenciaFechas)

	return nil
}

func horasExtras(file *excelize.File, novedad novedades.Novedades, row *int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("A%d", *row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("B%d", *row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("C%d", *row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("D%d", *row), recurso.Nombre)
	for _, recursoNovedad := range novedad.Recursos {
		file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("E%d", *row), recursoNovedad.Periodo)
		for _, horasExtrasNovedad := range recursoNovedad.HorasExtras {
			cell, ok := constantes.HorasExtrasTipos[(horasExtrasNovedad.TipoDia + horasExtrasNovedad.TipoHora)]
			if ok {
				file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("%s%d", cell, *row), horasExtrasNovedad.Cantidad)
			} else {
				fmt.Print("Error de tipo excel: " + (horasExtrasNovedad.TipoDia + horasExtrasNovedad.TipoHora))
			}
		}
		*row = *row + 1
	}
	return nil
}

func pagos(file *excelize.File, novedad novedades.Novedades, row *int) error {
	for _, recursoInterno := range novedad.Recursos {
		datosUsuario := strings.Split(recursoInterno.Recurso, "(")
		if len(datosUsuario) == 2 {
			datosUsuario[1] = strings.ReplaceAll(datosUsuario[1], ")", "")
			legajo, err := strconv.Atoi(datosUsuario[1])
			if err != nil {
				return err
			}
			err, recurso := recursos.GetRecursoInterno("", 0, legajo)
			if err != nil {
				return err
			}
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", *row), novedad.IdSecuencial)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", *row), novedad.Descripcion)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", *row), recurso.Legajo)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", *row), recurso.Nombre)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", *row), recurso.Apellido)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", *row), novedad.ImporteTotal)
			if strings.Contains(novedad.Descripcion, "retroactivo") {
				file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("G%d", *row), "SI")
			}
			*row = *row + 1
		}

	}

	*row = *row - 1

	return nil
}

func allNovedades(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}

	file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("A%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("B%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("C%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("D%d", row), recurso.Apellido)

	return nil
}

func initializeExcel(file *excelize.File) error {
	// General
	// Unir celdas de general para sueldos y prestamos
	file.MergeCell(constantes.PestanaGeneral, "F1", "H1")
	file.MergeCell(constantes.PestanaGeneral, "I1", "K1")

	// Ingresar los nombres de las celdas en general
	file.SetCellValue(constantes.PestanaGeneral, "A2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaGeneral, "B2", "TIPO DE NOVEDAD")
	file.SetCellValue(constantes.PestanaGeneral, "C2", "LEGAJO")
	file.SetCellValue(constantes.PestanaGeneral, "D2", "APELLIDO")
	file.SetCellValue(constantes.PestanaGeneral, "E2", "NOMBRE")
	file.SetCellValue(constantes.PestanaGeneral, "F2", "NUEVO SUELDO")
	file.SetCellValue(constantes.PestanaGeneral, "G2", "AJUSTE RETROACTIVIDAD")
	file.SetCellValue(constantes.PestanaGeneral, "H2", "DEVENGAMIENTO")
	file.SetCellValue(constantes.PestanaGeneral, "I2", "MONTO TOTAL")
	file.SetCellValue(constantes.PestanaGeneral, "J2", "MONTO CUOTA")
	file.SetCellValue(constantes.PestanaGeneral, "K2", "TOTAL CUOTAS")
	file.SetCellValue(constantes.PestanaGeneral, "L2", "MONTO GIMNASIO")
	file.SetCellValue(constantes.PestanaGeneral, "M2", "MONTO IDIOMA")
	file.SetCellValue(constantes.PestanaGeneral, "N2", "MONTO TARJETA")
	file.SetCellValue(constantes.PestanaGeneral, "F1", "NUEVOS SUELDOS")
	file.SetCellValue(constantes.PestanaGeneral, "I1", "ANTICIPO / PRESTAMO")
	file.SetCellValue(constantes.PestanaGeneral, "L1", "GYM")
	file.SetCellValue(constantes.PestanaGeneral, "M1", "IDIOMA")
	file.SetCellValue(constantes.PestanaGeneral, "N1", "TARJETA BENEFICIO")

	// Horas extras
	// Unir celdas de horas extras para sueldos y prestamos
	file.MergeCell(constantes.PestanaHorasExtras, "E1", "I1")

	// Ingresar los nombres de las celdas en horas extras
	file.SetCellValue(constantes.PestanaHorasExtras, "A2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaHorasExtras, "B2", "LEGAJO")
	file.SetCellValue(constantes.PestanaHorasExtras, "C2", "APELLIDO")
	file.SetCellValue(constantes.PestanaHorasExtras, "D2", "NOMBRE")
	file.SetCellValue(constantes.PestanaHorasExtras, "E2", "PERIODO")
	file.SetCellValue(constantes.PestanaHorasExtras, "F2", "AL 50% EXENTAS (CONCEPTO 2212)")
	file.SetCellValue(constantes.PestanaHorasExtras, "G2", "AL 100% EXENTAS (CONCEPTO 2220)")
	file.SetCellValue(constantes.PestanaHorasExtras, "H2", "HORAS FERIADO")
	file.SetCellValue(constantes.PestanaHorasExtras, "I2", "NOCTURNAS AL 50% (CONCEPTO 2213)")
	file.SetCellValue(constantes.PestanaHorasExtras, "J2", "NOCTURNAS AL 100% (CONCEPTO 2221)")

	// Licencias
	// Unir celdas de licencias para sueldos y prestamos
	file.MergeCell(constantes.PestanaLicencias, "E1", "G1")

	// Ingresar los nombres de las celdas en licencias
	file.SetCellValue(constantes.PestanaLicencias, "A2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaLicencias, "B2", "LEGAJO")
	file.SetCellValue(constantes.PestanaLicencias, "C2", "APELLIDO")
	file.SetCellValue(constantes.PestanaLicencias, "D2", "NOMBRE")
	file.SetCellValue(constantes.PestanaLicencias, "E2", "TIPO")
	file.SetCellValue(constantes.PestanaLicencias, "F2", "DIAS")

	// Todas las Novedades
	file.SetCellValue(constantes.PestanaNovedades, "A2", "ID")
	file.SetCellValue(constantes.PestanaNovedades, "B2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaNovedades, "C2", "NOMBRE")
	file.SetCellValue(constantes.PestanaNovedades, "D2", "APELLIDO")
	return nil
}

func verificacionNovedad(novedad novedades.Novedades, fechaDesde string, fechaHasta string) bool {
	if novedad.Fecha != "" {
		fechaNovedad, err := time.Parse(constantes.FormatoFechaProvicional, novedad.Fecha)
		if err != nil {
			return false
		}
		if fechaDesde != "" {
			fechaDesdeTime, err := time.Parse(constantes.FormatoFechaProvicional, fechaDesde)
			if err != nil {
				return false
			}
			if fechaDesdeTime.After(fechaNovedad) {
				return false
			}
		}
		if fechaHasta != "" {
			fechaHastaTime, err := time.Parse(constantes.FormatoFechaProvicional, fechaHasta)
			if err != nil {
				return false
			}
			if fechaHastaTime.Before(fechaNovedad) {
				return false
			}
		}
	}
	return true
}

// Crear excel
func GetExcelPP(c *fiber.Ctx) error {
	fmt.Println("GetExcelFile")

	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	var novedades []novedades.Novedades

	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)

	err := coll.FindOne(context.TODO(), bson.M{"tipo": "PP"}).Decode(&novedad)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("novedad no encontrada")
	}

	// cursor, err := coll.Find(context.TODO(), bson.M{})

	err = ExcelPP(novedades, c.Query("fechaDesde"), c.Query("fechaHasta"))

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error al crear archivo: " + err.Error())
	}
	return c.SendFile(os.Getenv("EXCELPP_FILE"))
}

// ingresar datos a un excel
func ExcelPP(novedadesArr []novedades.Novedades, fechaDesde string, fechaHasta string) error {

	// Abrir archivo de excel
	os.Remove(os.Getenv("EXCELPP_FILE"))
	file := excelize.NewFile()
	file.SetSheetName("Sheet1", constantes.PestanaPagoProvedores)

	initializeExcel(file)

	// guardar archivo
	err := file.SaveAs(os.Getenv("EXCELPP_FILE"))
	if err != nil {
		log.Printf("No se pudo guardar el archivo de Excel por el error %s", err.Error())
		return err
	}
	return nil
}
