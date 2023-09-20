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
	"github.com/proyectoNovedades/servicios/proveedores"
	"github.com/proyectoNovedades/servicios/recursos"
	"github.com/proyectoNovedades/servicios/userGoogle"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var novedad novedades.Novedades
var proveedor proveedores.Proveedores

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

	if c.Query("periodo") != "" {
		periodo := bson.D{{Key: "periodo", Value: c.Query("periodo")}}
		filter = bson.M{"$and": bson.A{usuarioExist, usuarioNotEmpty, descripcionExist, descripcionNotEmpty, EstadoNoRechazado, periodo}}
	}

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
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", *row), novedad.Fecha)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", *row), novedad.IdSecuencial)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", *row), novedad.Descripcion)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", *row), recurso.Legajo)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", *row), recurso.Nombre)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", *row), recurso.Apellido)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("G%d", *row), recursoInterno.Importe)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("P%d", *row), novedad.Periodo)
			if strings.Contains(novedad.Descripcion, "retroactivo") {
				file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("H%d", *row), "SI")
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
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", *row), novedad.Fecha)
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", *row), novedad.IdSecuencial)
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", *row), novedad.Descripcion)
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", *row), recurso.Legajo)
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", *row), distribucion.Cecos.Cliente)
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", *row), "")
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("G%d", *row), fmt.Sprintf("%v", distribucion.Porcentaje)+"%")
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("P%d", *row), novedad.Periodo)
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
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.Fecha)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("J%d", row), novedad.ImporteTotal)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("K%d", row), novedad.ImporteTotal/float64(cuotas))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("L%d", row), cuotas)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("P%d", row), novedad.Periodo)
	return nil
}

func gimnasio(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.Fecha)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("M%d", row), novedad.ImporteTotal)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("P%d", row), novedad.Periodo)

	return nil
}

func idioma(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.Fecha)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("N%d", row), novedad.ImporteTotal)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("P%d", row), novedad.Periodo)

	return nil
}

func tarjetaBeneficio(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.Fecha)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("O%d", row), novedad.ImporteTotal)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("P%d", row), novedad.Periodo)

	return nil
}

func licencias(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}
	diferenciaFechas, _ := strconv.Atoi(novedad.Cantidad)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("A%d", row), novedad.Fecha)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("B%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("C%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("D%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("E%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("F%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("G%d", row), diferenciaFechas)
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("H%d", row), novedad.Periodo)

	return nil
}

func horasExtras(file *excelize.File, novedad novedades.Novedades, row *int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("A%d", *row), recurso.Fecha)
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("B%d", *row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("C%d", *row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("D%d", *row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("E%d", *row), recurso.Apellido)
	for _, recursoNovedad := range novedad.Recursos {
		file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("F%d", *row), recursoNovedad.Periodo)
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
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", *row), novedad.Fecha)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", *row), novedad.IdSecuencial)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", *row), novedad.Descripcion)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", *row), recurso.Legajo)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", *row), recurso.Nombre)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", *row), recurso.Apellido)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("G%d", *row), novedad.ImporteTotal)
			if strings.Contains(novedad.Descripcion, "retroactivo") {
				file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("H%d", *row), "SI")
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

	file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("A%d", row), novedad.Fecha)
	file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("B%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("C%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("D%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("E%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("F%d", row), novedad.Periodo)

	return nil
}

func initializeExcel(file *excelize.File) error {
	// General
	// Unir celdas de general para sueldos y prestamos
	file.MergeCell(constantes.PestanaGeneral, "G1", "I1")
	file.MergeCell(constantes.PestanaGeneral, "J1", "L1")
	// Ingresar los nombres de las celdas en general
	file.SetCellValue(constantes.PestanaGeneral, "A2", "FECHA")
	file.SetCellValue(constantes.PestanaGeneral, "B2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaGeneral, "C2", "TIPO DE NOVEDAD")
	file.SetCellValue(constantes.PestanaGeneral, "D2", "LEGAJO")
	file.SetCellValue(constantes.PestanaGeneral, "E2", "NOMBRE")
	file.SetCellValue(constantes.PestanaGeneral, "F2", "APELLIDO")
	file.SetCellValue(constantes.PestanaGeneral, "G2", "NUEVO SUELDO")
	file.SetCellValue(constantes.PestanaGeneral, "H2", "AJUSTE RETROACTIVIDAD")
	file.SetCellValue(constantes.PestanaGeneral, "I2", "DEVENGAMIENTO")
	file.SetCellValue(constantes.PestanaGeneral, "J2", "MONTO TOTAL")
	file.SetCellValue(constantes.PestanaGeneral, "K2", "MONTO CUOTA")
	file.SetCellValue(constantes.PestanaGeneral, "L2", "TOTAL CUOTAS")
	file.SetCellValue(constantes.PestanaGeneral, "M2", "MONTO GIMNASIO")
	file.SetCellValue(constantes.PestanaGeneral, "N2", "MONTO IDIOMA")
	file.SetCellValue(constantes.PestanaGeneral, "O2", "MONTO TARJETA")
	file.SetCellValue(constantes.PestanaGeneral, "P2", "PERIODO")
	file.SetCellValue(constantes.PestanaGeneral, "G1", "NUEVOS SUELDOS")
	file.SetCellValue(constantes.PestanaGeneral, "J1", "ANTICIPO / PRESTAMO")
	file.SetCellValue(constantes.PestanaGeneral, "M1", "GYM")
	file.SetCellValue(constantes.PestanaGeneral, "N1", "IDIOMA")
	file.SetCellValue(constantes.PestanaGeneral, "O1", "TARJETA BENEFICIO")
	// Horas extras
	// Unir celdas de horas extras para sueldos y prestamos
	file.MergeCell(constantes.PestanaHorasExtras, "F1", "K1")
	// Ingresar los nombres de las celdas en horas extras
	file.SetCellValue(constantes.PestanaHorasExtras, "A2", "FECHA")
	file.SetCellValue(constantes.PestanaHorasExtras, "B2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaHorasExtras, "C2", "LEGAJO")
	file.SetCellValue(constantes.PestanaHorasExtras, "D2", "NOMBRE")
	file.SetCellValue(constantes.PestanaHorasExtras, "E2", "APELLIDO")
	file.SetCellValue(constantes.PestanaHorasExtras, "F2", "PERIODO")
	file.SetCellValue(constantes.PestanaHorasExtras, "G2", "AL 50% EXENTAS (CONCEPTO 2212)")
	file.SetCellValue(constantes.PestanaHorasExtras, "H2", "AL 100% EXENTAS (CONCEPTO 2220)")
	file.SetCellValue(constantes.PestanaHorasExtras, "I2", "NOCTURNAS AL 50% (CONCEPTO 2213)")
	file.SetCellValue(constantes.PestanaHorasExtras, "J2", "NOCTURNAS AL 100% (CONCEPTO 2221)")
	file.SetCellValue(constantes.PestanaHorasExtras, "K2", "HORAS FERIADO")
	file.SetCellValue(constantes.PestanaHorasExtras, "F1", "")
	// Licencias
	// Unir celdas de licencias para sueldos y prestamos
	file.MergeCell(constantes.PestanaLicencias, "F1", "G1")
	// Ingresar los nombres de las celdas en licencias
	file.SetCellValue(constantes.PestanaLicencias, "A2", "FECHA")
	file.SetCellValue(constantes.PestanaLicencias, "B2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaLicencias, "C2", "LEGAJO")
	file.SetCellValue(constantes.PestanaLicencias, "D2", "NOMBRE")
	file.SetCellValue(constantes.PestanaLicencias, "E2", "APELLIDO")
	file.SetCellValue(constantes.PestanaLicencias, "F2", "TIPO")
	file.SetCellValue(constantes.PestanaLicencias, "G2", "DIAS")
	file.SetCellValue(constantes.PestanaLicencias, "H2", "PERIODO")
	file.SetCellValue(constantes.PestanaLicencias, "F1", "")
	// Todas las Novedades
	file.SetCellValue(constantes.PestanaNovedades, "A2", "FECHA")
	file.SetCellValue(constantes.PestanaNovedades, "B2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaNovedades, "C2", "TIPO DE NOVEDAD")
	file.SetCellValue(constantes.PestanaNovedades, "D2", "NOMBRE")
	file.SetCellValue(constantes.PestanaNovedades, "E2", "APELLIDO")
	file.SetCellValue(constantes.PestanaNovedades, "F2", "PERIODO")
	// Ingresar los nombres de las celdas en proveedores
	file.SetCellValue(constantes.PestanaPagoProvedores, "A2", "FECHA")
	file.SetCellValue(constantes.PestanaPagoProvedores, "B2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaPagoProvedores, "C2", "ESTADO")
	file.SetCellValue(constantes.PestanaPagoProvedores, "D2", "PROVEEDOR")
	file.SetCellValue(constantes.PestanaPagoProvedores, "E2", "COD PROV")
	file.SetCellValue(constantes.PestanaPagoProvedores, "F2", "RAZON SOCIAL")
	file.SetCellValue(constantes.PestanaPagoProvedores, "G2", "IMPORTE TOTAL")
	file.SetCellValue(constantes.PestanaPagoProvedores, "H2", "USUARIO")
	file.SetCellValue(constantes.PestanaPagoProvedores, "I2", "COMENTARIOS")
	file.SetCellValue(constantes.PestanaPagoProvedores, "J2", "PERIODO")
	// Ingresar los nombres de las celdas en facturacion de servicios
	file.SetCellValue(constantes.PestanaFactServicios, "A2", "FECHA")
	file.SetCellValue(constantes.PestanaFactServicios, "B2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaFactServicios, "C2", "TIPO DE NOVEDAD")
	file.SetCellValue(constantes.PestanaFactServicios, "D2", "LEGAJO")
	file.SetCellValue(constantes.PestanaFactServicios, "E2", "NOMBRE")
	file.SetCellValue(constantes.PestanaFactServicios, "F2", "APELLIDO")
	file.SetCellValue(constantes.PestanaFactServicios, "G2", "IMPORTE TARJETA")
	file.SetCellValue(constantes.PestanaFactServicios, "H2", "PERIODO")

	// Ingresar los nombres de las celdas en rendicion de costos
	file.SetCellValue(constantes.PestanaRendCostos, "A2", "FECHA")
	file.SetCellValue(constantes.PestanaRendCostos, "B2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaRendCostos, "C2", "TIPO DE NOVEDAD")
	file.SetCellValue(constantes.PestanaRendCostos, "D2", "LEGAJO")
	file.SetCellValue(constantes.PestanaRendCostos, "E2", "NOMBRE")
	file.SetCellValue(constantes.PestanaRendCostos, "F2", "APELLIDO")
	file.SetCellValue(constantes.PestanaRendCostos, "G2", "IMPORTE TARJETA")
	file.SetCellValue(constantes.PestanaRendCostos, "H2", "PERIODO")

	// Ingresar los nombres de las celdas en nuevo ceco
	file.SetCellValue(constantes.PestanaNuevoCeco, "A2", "FECHA")
	file.SetCellValue(constantes.PestanaNuevoCeco, "B2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaNuevoCeco, "C2", "TIPO DE NOVEDAD")
	file.SetCellValue(constantes.PestanaNuevoCeco, "D2", "LEGAJO")
	file.SetCellValue(constantes.PestanaNuevoCeco, "E2", "NOMBRE")
	file.SetCellValue(constantes.PestanaNuevoCeco, "F2", "APELLIDO")
	file.SetCellValue(constantes.PestanaNuevoCeco, "G2", "IMPORTE TARJETA")
	file.SetCellValue(constantes.PestanaNuevoCeco, "H2", "PERIODO")

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
	err, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if err != nil {
		return c.Status(codigo).SendString(err.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionNovedad)
	err = coll.FindOne(context.TODO(), bson.M{"tipo": "PP"}).Decode(&novedad)

	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("novedad no encontrada")
	}

	var novedades []novedades.Novedades

	cursor, err := coll.Find(context.TODO(), bson.M{})

	if err = cursor.All(context.Background(), &novedades); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

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
	var rowPagoProveedores int = 3
	for _, item := range novedadesArr {
		if !verificacionNovedad(item, fechaDesde, fechaHasta) {
			continue
		}
		var pasosWorkflow novedades.PasosWorkflow
		coll := client.Database(constantes.Database).Collection(constantes.CollectionPasosWorkflow)
		err := coll.FindOne(context.TODO(), bson.M{"tipo": item.Descripcion}).Decode(&pasosWorkflow)
		err = pagoProveedores(file, item, rowPagoProveedores)
		if err == nil {
			rowPagoProveedores++
		}
	}
	// guardar archivo
	if err := file.SaveAs(os.Getenv("EXCELPP_FILE")); err != nil {
		log.Printf("No se pudo guardar el archivo de Excel: %s", err.Error())
		return err
	}
	return nil

}

func GetExcelAdmin(c *fiber.Ctx) error {
	fmt.Println("GetExcelFile")

	// validar el token
	err, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.Admin)
	if err != nil {
		return c.Status(codigo).SendString(err.Error())
	}
	//buscar todo en mongo y filtrarlo por los datos que necesitamos validados
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

	if c.Query("periodo") != "" {
		periodo := bson.D{{Key: "periodo", Value: c.Query("periodo")}}
		filter = bson.M{"$and": bson.A{usuarioExist, usuarioNotEmpty, descripcionExist, descripcionNotEmpty, EstadoNoRechazado, periodo}}
	}

	cursor, err = coll.Find(context.TODO(), filter, opts)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	var novedades []novedades.Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	err = excelAdmin(novedades, c.Query("fechaDesde"), c.Query("fechaHasta"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error al crear archivo: " + err.Error())
	}
	return c.SendFile(os.Getenv("EXCEL_FILE_ADMIN"))
}

// ingresar datos a un excel
func excelAdmin(novedadesArr []novedades.Novedades, fechaDesde string, fechaHasta string) error {

	// Abrir archivo de excel
	os.Remove(os.Getenv("EXCEL_FILE_ADMIN"))
	file := excelize.NewFile()
	file.SetSheetName("Sheet1", constantes.PestanaFactServicios)
	file.NewSheet(constantes.PestanaRendCostos)
	file.NewSheet(constantes.PestanaNuevoCeco)
	file.NewSheet(constantes.PestanaHorasExtras)
	initializeExcel(file)
	var rowFactServicios int = 3
	var rowRendCostos int = 3
	var rowNuevoCeco int = 3
	var rowHorasExtras int = 3

	for _, item := range novedadesArr {
		if !verificacionNovedad(item, fechaDesde, fechaHasta) {
			continue
		}
		var pasosWorkflow novedades.PasosWorkflow
		coll := client.Database(constantes.Database).Collection(constantes.CollectionPasosWorkflow)
		err := coll.FindOne(context.TODO(), bson.M{"tipo": item.Descripcion}).Decode(&pasosWorkflow)

		if pasosWorkflow.TipoExcel == constantes.FactServicios {
			err = factServicios(file, item, rowFactServicios)
			if err == nil {
				rowFactServicios++
			}
		}
		if pasosWorkflow.TipoExcel == constantes.RendCostos {
			err = rendCostos(file, item, rowRendCostos)
			if err == nil {
				rowRendCostos++
			}
		}
		if pasosWorkflow.TipoExcel == constantes.DescHorasExtras {
			err = horasExtras(file, item, &rowHorasExtras)
			if err == nil {
				rowHorasExtras++
			}
		}
		err = nuevoCeco(file, item, rowNuevoCeco)
		if err == nil {
			rowNuevoCeco++
		}

	}
	// guardar archivo
	err := file.SaveAs(os.Getenv("EXCEL_FILE_ADMIN"))
	if err != nil {
		log.Printf("No se pudo guardar el archivo de Excel por el error %s", err.Error())
		return err
	}
	return nil
}

func factServicios(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}
	cellMappings := map[string]interface{}{
		fmt.Sprintf("A%d", row): novedad.Fecha,
		fmt.Sprintf("B%d", row): novedad.IdSecuencial,
		fmt.Sprintf("C%d", row): novedad.Descripcion,
		fmt.Sprintf("D%d", row): recurso.Legajo,
		fmt.Sprintf("E%d", row): recurso.Nombre,
		fmt.Sprintf("F%d", row): recurso.Apellido,
		fmt.Sprintf("N%d", row): novedad.ImporteTotal,
		fmt.Sprintf("P%d", row): novedad.Periodo,
	}
	for cell, value := range cellMappings {
		file.SetCellValue(constantes.PestanaFactServicios, cell, value)
	}
	return nil
}
func rendCostos(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}
	cellMappings := map[string]interface{}{
		fmt.Sprintf("A%d", row): novedad.Fecha,
		fmt.Sprintf("B%d", row): novedad.IdSecuencial,
		fmt.Sprintf("C%d", row): novedad.Descripcion,
		fmt.Sprintf("D%d", row): recurso.Legajo,
		fmt.Sprintf("E%d", row): recurso.Nombre,
		fmt.Sprintf("F%d", row): recurso.Apellido,
		fmt.Sprintf("N%d", row): novedad.ImporteTotal,
		fmt.Sprintf("P%d", row): novedad.Periodo,
	}
	for cell, value := range cellMappings {
		file.SetCellValue(constantes.PestanaRendCostos, cell, value)
	}

	return nil
}
func nuevoCeco(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}

	cellMappings := map[string]interface{}{
		fmt.Sprintf("A%d", row): novedad.Fecha,
		fmt.Sprintf("B%d", row): novedad.IdSecuencial,
		fmt.Sprintf("C%d", row): novedad.Descripcion,
		fmt.Sprintf("D%d", row): recurso.Legajo,
		fmt.Sprintf("E%d", row): recurso.Nombre,
		fmt.Sprintf("F%d", row): recurso.Apellido,
		fmt.Sprintf("N%d", row): novedad.ImporteTotal,
		fmt.Sprintf("P%d", row): novedad.Periodo,
	}
	for cell, value := range cellMappings {
		file.SetCellValue(constantes.PestanaNuevoCeco, cell, value)
	}

	return nil
}

func pagoProveedores(file *excelize.File, novedad novedades.Novedades, row int) error {
	proveedor, err := proveedores.ObtenerProveedores(0, 0, novedad.Proveedor)
	if err != nil {
		// Maneja el error si es necesario
		return err
	}

	cellMappings := map[string]interface{}{
		fmt.Sprintf("A%d", row): novedad.Fecha,
		fmt.Sprintf("B%d", row): novedad.IdSecuencial,
		fmt.Sprintf("C%d", row): novedad.Estado,
		fmt.Sprintf("D%d", row): novedad.Proveedor,
		fmt.Sprintf("E%d", row): proveedor.CodProv,
		fmt.Sprintf("F%d", row): proveedor.RazonSocial,
		fmt.Sprintf("G%d", row): novedad.ImporteTotal,
		fmt.Sprintf("H%d", row): novedad.Usuario,
		fmt.Sprintf("I%d", row): novedad.Comentarios,
		fmt.Sprintf("J%d", row): novedad.Periodo,
	}

	for cell, value := range cellMappings {
		file.SetCellValue(constantes.PestanaPagoProvedores, cell, value)
	}

	return nil
}
