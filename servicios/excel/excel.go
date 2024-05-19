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
	"github.com/proyectoNovedades/servicios/models"
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

// Crear excel People Operation
func GetExcelFile(c *fiber.Ctx) error {
	fmt.Println("GetExcelFilePeopleOperation")

	fechaDesde := c.Query("fechaDesde")
	fechaHasta := c.Query("fechaHasta")

	// Validar token
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
	TiposDePeopleOperation := bson.D{{Key: "tipo", Value: bson.M{"$regex": "RH|NP|PB"}}}
	IgnorarDescripcion := bson.D{{Key: "descripcion", Value: bson.M{"$ne": "Alta beneficios"}}}

	filter := bson.M{"$and": bson.A{usuarioExist, usuarioNotEmpty, descripcionExist, descripcionNotEmpty, EstadoNoRechazado, TiposDePeopleOperation, IgnorarDescripcion}}
	opts := options.Find().SetSort(bson.D{{Key: "descripcion", Value: 1}, {Key: "usuario", Value: 1}})

	if c.Query("periodo") != "" {
		var periodoFormato string
		if len(c.Query("periodo")) == 6 {
			periodoFormato = "0" + c.Query("periodo")
		} else {
			periodoFormato = c.Query("periodo")
		}
		periodoFormatoFechaDesde, err := time.Parse("01-2006", periodoFormato)
		periodoFormatoFechaHasta := periodoFormatoFechaDesde.AddDate(0, 1, 0)
		if err == nil {
			fechaDesde = periodoFormatoFechaDesde.Format(constantes.FormatoFechaProvicional)
			fechaHasta = periodoFormatoFechaHasta.Format(constantes.FormatoFechaProvicional)
			fmt.Println(fechaDesde, fechaHasta)
		} else {
			fmt.Println(err)
		}
		periodo := bson.D{{Key: "$or", Value: bson.A{bson.D{{Key: "periodo", Value: c.Query("periodo")}}, bson.D{{Key: "periodo", Value: ""}}}}}
		filter = bson.M{"$and": bson.A{usuarioExist, usuarioNotEmpty, descripcionExist, descripcionNotEmpty, EstadoNoRechazado, TiposDePeopleOperation, IgnorarDescripcion, periodo}}
	}

	cursor, err = coll.Find(context.TODO(), filter, opts)

	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var novedades []novedades.Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	fmt.Print(len(novedades))
	fmt.Println(" novedades found")

	seBuscaPorPeriodo := c.Query("periodo") != ""

	err = datosExcel(novedades, fechaDesde, fechaHasta, seBuscaPorPeriodo)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error al crear archivo: " + err.Error())
	}
	return c.SendFile(os.Getenv("EXCEL_FILE"))
}

// Crear excel PP
func GetExcelPP(c *fiber.Ctx) error {
	fmt.Println("GetExcelFilePP")

	// Validar token
	err, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.Admin)
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

// Crear excel Administration
func GetExcelAdmin(c *fiber.Ctx) error {
	fmt.Println("GetExcelFileAdministration")

	// Validar token
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

// Crear excel Freelance
func GetFreelancesListExcel(c *fiber.Ctx) error {
	fmt.Println("GetFreelancesListExcel")

	// Validar token
	err, codigo, email := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if err != nil {
		return c.Status(codigo).SendString(err.Error())
	}

	err, userGoogle := userGoogle.GetInternUserITP(email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	coll := client.Database(constantes.Database).Collection(constantes.CollectionFreelance)
	filter := bson.D{}
	if !stringContains(constantes.BoardAndPO, userGoogle.Rol) {
		filter = bson.D{{Key: "vertical", Value: userGoogle.Rol}}
	}
	opts := options.Find().SetSort(bson.D{{Key: "idFreelance", Value: 1}})
	cursor, _ := coll.Find(context.Background(), filter, opts)
	var results []models.Freelances
	err = cursor.All(context.Background(), &results)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	err = ingresarDatosExcelFreelance(results)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error al crear archivo: " + err.Error())
	}
	return c.SendFile(os.Getenv("EXCEL_FILE_FREELANCE"))
}

// Ingresar datos a un excel
func datosExcel(novedadesArr []novedades.Novedades, fechaDesde string, fechaHasta string, periodo bool) error {

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

	//obtener todos los pasos de workflow
	coll := client.Database(constantes.Database).Collection(constantes.CollectionPasosWorkflow)
	cursor, err := coll.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	var pasosWorkflow []novedades.PasosWorkflow
	if err = cursor.All(context.Background(), &pasosWorkflow); err != nil {
		return err
	}

	// crear un mapa strins -> string con key descripcion y value tipoExcel
	var pasosWorkflowMap map[string]string = make(map[string]string)
	for _, item := range pasosWorkflow {
		pasosWorkflowMap[item.Tipo] = item.TipoExcel
	}

	for _, item := range novedadesArr {
		if !verificacionNovedad(item, fechaDesde, fechaHasta) {
			if periodo {
				if item.Periodo == "" {
					continue
				}
			} else {
				continue
			}
		}

		if pasosWorkflowMap[item.Descripcion] == constantes.DescAnticipo {
			err = anticipoPrestamo(file, item, rowGeneral, 1)
			if err == nil {
				rowGeneral = rowGeneral + 1
			}
		}
		if pasosWorkflowMap[item.Descripcion] == constantes.DescSueldoNuevo {
			err = nuevoSueldo(file, item, &rowGeneral)
		}
		if pasosWorkflowMap[item.Descripcion] == constantes.DescSueldoNuevoMasivo {
			err = nuevoSueldoMasivo(file, item, &rowGeneral)
		}
		if pasosWorkflowMap[item.Descripcion] == constantes.DescLicencia {
			err = licencias(file, item, rowLicencias)
			if err == nil {
				rowLicencias = rowLicencias + 1
			}
		}
		if pasosWorkflowMap[item.Descripcion] == constantes.DescPrestamo {
			err = anticipoPrestamo(file, item, rowGeneral, 6)
			if err == nil {
				rowGeneral = rowGeneral + 1
			}
		}
		if pasosWorkflowMap[item.Descripcion] == constantes.DescTarjetaBeneficios {
			err = tarjetaBeneficio(file, item, rowGeneral)
			if err == nil {
				rowGeneral = rowGeneral + 1
			}
		}
		if pasosWorkflowMap[item.Descripcion] == constantes.DescHorasExtras {
			horasExtras(file, item, &rowHorasExtras)
		}
		if pasosWorkflowMap[item.Descripcion] == constantes.DescPagos {
			pagos(file, item, &rowGeneral)
		}
		err, rowNovedades = allNovedades(file, item, rowNovedades)
		if err == nil {
			rowNovedades = rowNovedades + 1
		}
	}

	// Guardar archivo
	err = file.SaveAs(os.Getenv("EXCEL_FILE"))
	if err != nil {
		log.Printf("No se pudo guardar el archivo de Excel por el error %s", err.Error())
		return err
	}
	return nil
}

// Ingresar datos a un excel
func ingresarDatosExcelFreelance(freelancesList []models.Freelances) error {

	// Abrir archivo de excel
	os.Remove("EXCEL_FILE_FREELANCE")
	file := excelize.NewFile()
	file.SetSheetName("Sheet1", constantes.PestanaGeneral)

	err := initializeExcelFreelances(file)
	if err != nil {
		return err
	}

	var RowGeneral int = 3
	for _, item := range freelancesList {
		err, rows := freelanceInsert(file, item, RowGeneral)
		if err == nil {
			RowGeneral = RowGeneral + 1 + rows
		}
	}

	// Guardar archivo
	err = file.SaveAs(os.Getenv("EXCEL_FILE_FREELANCE"))
	if err != nil {
		log.Printf("No se pudo guardar el archivo de Excel por el error %s", err.Error())
		return err
	}
	return nil
}

// Ingresar datos a un excel
func ExcelPP(novedadesArr []novedades.Novedades, fechaDesde string, fechaHasta string) error {

	// Abrir archivo de excel
	os.Remove(os.Getenv("EXCELPP_FILE"))
	file := excelize.NewFile()
	file.SetSheetName("Sheet1", constantes.PestanaPagoProvedores)
	initializeExcelAdmin(file)
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

	// Guardar archivo
	if err := file.SaveAs(os.Getenv("EXCELPP_FILE")); err != nil {
		log.Printf("No se pudo guardar el archivo de Excel: %s", err.Error())
		return err
	}
	return nil
}

// Ingresar datos a un excel
func excelAdmin(novedadesArr []novedades.Novedades, fechaDesde string, fechaHasta string) error {

	// Abrir archivo de excel
	os.Remove(os.Getenv("EXCEL_FILE_ADMIN"))
	file := excelize.NewFile()
	file.SetSheetName("Sheet1", constantes.PestanaFactServicios)
	file.NewSheet(constantes.PestanaRendGastos)
	file.NewSheet(constantes.PestanaNuevoCeco)
	file.NewSheet(constantes.PestanaHorasExtras)
	initializeExcelAdmin(file)
	var rowFactServicios int = 3
	var rowRendGastos int = 3
	var rowNuevoCeco int = 3
	var rowHorasExtras int = 3

	//obtener todos los pasos de workflow
	coll := client.Database(constantes.Database).Collection(constantes.CollectionPasosWorkflow)
	cursor, err := coll.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	var pasosWorkflow []novedades.PasosWorkflow
	if err = cursor.All(context.Background(), &pasosWorkflow); err != nil {
		return err
	}

	// crear un mapa strins -> string con key descripcion y value tipoExcel
	var pasosWorkflowMap map[string]string = make(map[string]string)
	for _, item := range pasosWorkflow {
		pasosWorkflowMap[item.Tipo] = item.TipoExcel
	}

	// recorrer las novedades y guardarlas en el excel
	for _, item := range novedadesArr {
		if !verificacionNovedad(item, fechaDesde, fechaHasta) {
			continue
		}

		if pasosWorkflowMap[item.Descripcion] == constantes.DescFactServicios {
			err = factServicios(file, item, rowFactServicios)
			if err == nil {
				rowFactServicios++
			}
		}
		if pasosWorkflowMap[item.Descripcion] == constantes.DescRendGastos {
			quantityOfRowUsed, err := rendGastos(file, item, rowRendGastos)
			if err == nil {
				rowRendGastos += quantityOfRowUsed
			}
		}
		if pasosWorkflowMap[item.Descripcion] == constantes.DescHorasExtras {
			horasExtras(file, item, &rowHorasExtras)
		}
		if pasosWorkflowMap[item.Descripcion] == constantes.DescNuevoCeco {
			err = nuevoCeco(file, item, rowNuevoCeco)
			if err == nil {
				rowNuevoCeco++
			}
		}
	}
	// Guardar archivo
	err = file.SaveAs(os.Getenv("EXCEL_FILE_ADMIN"))
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
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", *row), utf8_decode(recurso.Nombre))
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", *row), utf8_decode(recurso.Apellido))
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("G%d", *row), recursoInterno.Importe)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("P%d", *row), novedad.Periodo)
			if strings.Contains(novedad.Descripcion, "retroactivo") {
				file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("H%d", *row), "SI")
			}
			*row = *row + 1
		}

	}

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
	return nil
}

func anticipoPrestamo(file *excelize.File, novedad novedades.Novedades, row int, cuotas float64) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), novedad.Fecha)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), utf8_decode(recurso.Nombre))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", row), utf8_decode(recurso.Apellido))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("J%d", row), novedad.ImporteTotal)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("K%d", row), novedad.ImporteTotal/cuotas)
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
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), utf8_decode(recurso.Nombre))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", row), utf8_decode(recurso.Apellido))
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
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), utf8_decode(recurso.Nombre))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", row), utf8_decode(recurso.Apellido))
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
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), utf8_decode(recurso.Nombre))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", row), utf8_decode(recurso.Apellido))
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
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("D%d", row), utf8_decode(recurso.Nombre))
	file.SetCellValue(constantes.PestanaLicencias, fmt.Sprintf("E%d", row), utf8_decode(recurso.Apellido))
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
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("A%d", *row), novedad.Fecha)
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("B%d", *row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("C%d", *row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("D%d", *row), utf8_decode(recurso.Nombre))
	file.SetCellValue(constantes.PestanaHorasExtras, fmt.Sprintf("E%d", *row), utf8_decode(recurso.Apellido))
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
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", *row), utf8_decode(recurso.Nombre))
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", *row), utf8_decode(recurso.Apellido))
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("G%d", *row), novedad.ImporteTotal)
			file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("P%d", *row), novedad.Periodo)
			if strings.Contains(novedad.Descripcion, "retroactivo") {
				file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("H%d", *row), "SI")
			}
			*row = *row + 1
		}

	}

	return nil
}

func allNovedades(file *excelize.File, novedad novedades.Novedades, row int) (error, int) {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err, row
	}
	if len(novedad.Recursos) == 0 {
		novedad.Recursos = append(novedad.Recursos, novedades.RecursosNovedades{Recurso: recurso.Nombre + " " + recurso.Apellido})
	}

	for _, recursoInterno := range novedad.Recursos {
		var rowStop int = row
		for index, distribucion := range novedad.Distribuciones {
			rowStop = row + index
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("A%d", rowStop), novedad.Fecha)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("B%d", rowStop), novedad.IdSecuencial)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("C%d", rowStop), novedad.Tipo)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("D%d", rowStop), novedad.Descripcion)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("E%d", rowStop), utf8_decode(recurso.Nombre))
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("F%d", rowStop), utf8_decode(recurso.Apellido))
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("G%d", rowStop), novedad.Periodo)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("H%d", rowStop), novedad.Fecha)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("I%d", rowStop), novedad.Hora)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("J%d", rowStop), novedad.FechaDesde)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("K%d", rowStop), novedad.FechaHasta)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("L%d", rowStop), novedad.Cantidad)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("M%d", rowStop), novedad.ImporteTotal)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("N%d", rowStop), novedad.Comentarios)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("O%d", rowStop), novedad.Estado)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("S%d", rowStop), distribucion.Cecos.Descripcion)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("T%d", rowStop), distribucion.Porcentaje)
		}
		if rowStop == row {
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("A%d", row), novedad.Fecha)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("B%d", row), novedad.IdSecuencial)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("C%d", row), novedad.Tipo)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("D%d", row), novedad.Descripcion)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("E%d", row), utf8_decode(recurso.Nombre))
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("F%d", row), utf8_decode(recurso.Apellido))
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("G%d", row), novedad.Periodo)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("H%d", row), novedad.Fecha)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("I%d", row), novedad.Hora)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("J%d", row), novedad.FechaDesde)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("K%d", row), novedad.FechaHasta)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("L%d", row), novedad.Cantidad)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("M%d", row), novedad.ImporteTotal)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("N%d", row), novedad.Comentarios)
			file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("O%d", row), novedad.Estado)
		}
		file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("P%d", row), recursoInterno.Recurso)
		file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("Q%d", row), recursoInterno.Importe)
		file.SetCellValue(constantes.PestanaNovedades, fmt.Sprintf("R%d", row), recursoInterno.Periodo)
		row = rowStop + 1
	}
	row--

	return nil, row
}

func freelanceInsert(file *excelize.File, freelance models.Freelances, row int) (error, int) {
	var quantityOfCecos int = 0
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("A%d", row), freelance.IdFreelance)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("B%d", row), freelance.NroFreelance)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("C%d", row), freelance.CUIT)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("D%d", row), utf8_decode(freelance.Nombre))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("E%d", row), utf8_decode(freelance.Apellido))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("F%d", row), freelance.FechaIngreso.Format(constantes.FormatoFecha))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("G%d", row), freelance.Nomina)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("H%d", row), freelance.Gerente)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("I%d", row), freelance.Vertical)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("J%d", row), freelance.HorasMen)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("K%d", row), freelance.Cargo)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("L%d", row), freelance.FacturaMonto)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("M%d", row), freelance.FacturaDesde.Format(constantes.FormatoFecha))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("N%d", row), freelance.FacturaADCuit)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("O%d", row), freelance.FacturaADMonto)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("P%d", row), freelance.FacturaADDesde.Format(constantes.FormatoFecha))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("Q%d", row), freelance.B21Monto)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("R%d", row), freelance.B21Desde.Format(constantes.FormatoFecha))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("S%d", row), freelance.Comentario)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("T%d", row), freelance.Habilitado)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("U%d", row), freelance.FechaBaja.Format(constantes.FormatoFecha))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("V%d", row), freelance.Telefono)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("W%d", row), freelance.EmailLaboral)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("X%d", row), freelance.EmailPersonal)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("Y%d", row), freelance.FechaNacimiento.Format(constantes.FormatoFecha))
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("Z%d", row), freelance.Genero)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("AA%d", row), freelance.Nacionalidad)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("AB%d", row), freelance.DomCalle)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("AC%d", row), freelance.DomNumero)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("AD%d", row), freelance.DomPiso)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("AE%d", row), freelance.DomDepto)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("AF%d", row), freelance.DomLocalidad)
	file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("AG%d", row), freelance.DomProvincia)
	for index, ceco := range freelance.Cecos {
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("AH%d", row+index), ceco.CcNum)
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("AI%d", row+index), ceco.CcCliente)
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("AJ%d", row+index), ceco.CcNombre)
		file.SetCellValue(constantes.PestanaGeneral, fmt.Sprintf("AK%d", row+index), ceco.CcPorcentaje)
		quantityOfCecos = index
	}

	return nil, quantityOfCecos
}

func factServicios(file *excelize.File, novedad novedades.Novedades, row int) error {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return err
	}
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("A%d", row), novedad.Fecha)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("B%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("C%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("D%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("E%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("F%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("G%d", row), novedad.Cliente)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("H%d", row), novedad.ConceptoDeFacturacion)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("I%d", row), novedad.Periodo)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("J%d", row), novedad.ImporteTotal)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("K%d", row), novedad.Estado)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("L%d", row), novedad.Motivo)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("M%d", row), novedad.Proveedor)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("N%d", row), novedad.Comentarios)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("O%d", row), novedad.OrdenDeCompra)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("P%d", row), novedad.Promovido)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("Q%d", row), novedad.EnviarA)
	file.SetCellValue(constantes.PestanaFactServicios, fmt.Sprintf("R%d", row), novedad.Contacto)
	return nil
}

func rendGastos(file *excelize.File, novedad novedades.Novedades, row int) (int, error) {
	err, recurso := recursos.GetRecursoInterno(novedad.Usuario, 0, 0)
	if err != nil {
		return 0, err
	}
	var quantityOfRowUsed int = 0
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("A%d", row), novedad.Fecha)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("B%d", row), novedad.IdSecuencial)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("C%d", row), novedad.Descripcion)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("D%d", row), recurso.Legajo)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("E%d", row), recurso.Nombre)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("F%d", row), recurso.Apellido)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("G%d", row), novedad.Proveedor)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("H%d", row), novedad.Contacto)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("I%d", row), novedad.ConceptoDeFacturacion)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("J%d", row), novedad.Periodo)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("K%d", row), novedad.ImporteTotal)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("L%d", row), novedad.Estado)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("M%d", row), novedad.Prioridad)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("N%d", row), novedad.Cliente)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("O%d", row), novedad.Comentarios)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("P%d", row), novedad.EnviarA)
	file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("Q%d", row), novedad.Contacto)
	for i, recursoInt := range novedad.Recursos {
		file.SetCellValue(constantes.PestanaRendGastos, fmt.Sprintf("R%d", row), recursoInt.Recurso)
		quantityOfRowUsed = i
	}
	quantityOfRowUsed++
	return quantityOfRowUsed, nil
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
		fmt.Sprintf("G%d", row): novedad.Proveedor,
		fmt.Sprintf("H%d", row): novedad.Contacto,
		fmt.Sprintf("I%d", row): novedad.ConceptoDeFacturacion,
		fmt.Sprintf("J%d", row): novedad.Periodo,
		fmt.Sprintf("K%d", row): novedad.ImporteTotal,
		fmt.Sprintf("L%d", row): novedad.Estado,
		fmt.Sprintf("M%d", row): novedad.Prioridad,
		fmt.Sprintf("N%d", row): novedad.Cliente,
		fmt.Sprintf("O%d", row): novedad.Comentarios,
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

func utf8_decode(str string) string {
	var result string
	for i := range str {
		result += string(str[i])
	}
	return result
}

func stringContains(str string, substr string) bool {
	if str != "" {
		result := strings.Split(str, ",")
		for _, part := range result {
			if part == substr {
				return true
			}
		}
	}
	return false
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
	file.SetCellValue(constantes.PestanaHorasExtras, "F2", "FECHA HORAS EXTRAS")
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
	file.SetCellValue(constantes.PestanaNovedades, "D2", "DESCRIPCION")
	file.SetCellValue(constantes.PestanaNovedades, "E2", "NOMBRE")
	file.SetCellValue(constantes.PestanaNovedades, "F2", "APELLIDO")
	file.SetCellValue(constantes.PestanaNovedades, "G2", "PERIODO")
	file.SetCellValue(constantes.PestanaNovedades, "H2", "FECHA")
	file.SetCellValue(constantes.PestanaNovedades, "I2", "HORA")
	file.SetCellValue(constantes.PestanaNovedades, "J2", "FECHA DESDE")
	file.SetCellValue(constantes.PestanaNovedades, "K2", "FECHA HASTA")
	file.SetCellValue(constantes.PestanaNovedades, "L2", "CANTIDAD")
	file.SetCellValue(constantes.PestanaNovedades, "M2", "IMPORTE TOTAL")
	file.SetCellValue(constantes.PestanaNovedades, "N2", "COMENTARIOS")
	file.SetCellValue(constantes.PestanaNovedades, "O2", "ESTADO")
	file.SetCellValue(constantes.PestanaNovedades, "P2", "RECURSO")
	file.SetCellValue(constantes.PestanaNovedades, "Q2", "IMPORTE")
	file.SetCellValue(constantes.PestanaNovedades, "R2", "PERIODO")
	file.SetCellValue(constantes.PestanaNovedades, "S2", "CENTRO DE COSTOS")
	file.SetCellValue(constantes.PestanaNovedades, "T2", "PORCENTAJE")
	return nil
}

func initializeExcelFreelances(file *excelize.File) error {
	file.SetCellValue(constantes.PestanaGeneral, "A2", "ID")
	file.SetCellValue(constantes.PestanaGeneral, "B2", "NRO FL")
	file.SetCellValue(constantes.PestanaGeneral, "C2", "CUIT")
	file.SetCellValue(constantes.PestanaGeneral, "D2", "NOMBRE")
	file.SetCellValue(constantes.PestanaGeneral, "E2", "APELLIDO")
	file.SetCellValue(constantes.PestanaGeneral, "F2", "FECHA DE INGRESO")
	file.SetCellValue(constantes.PestanaGeneral, "G2", "NOMINA")
	file.SetCellValue(constantes.PestanaGeneral, "H2", "GERENTE")
	file.SetCellValue(constantes.PestanaGeneral, "I2", "VERTICAL")
	file.SetCellValue(constantes.PestanaGeneral, "J2", "HORAS MEN")
	file.SetCellValue(constantes.PestanaGeneral, "K2", "CARGO")
	file.SetCellValue(constantes.PestanaGeneral, "L2", "FACTURA MONTO")
	file.SetCellValue(constantes.PestanaGeneral, "M2", "FACTURA DESDE")
	file.SetCellValue(constantes.PestanaGeneral, "N2", "FACTURA AD CUIT")
	file.SetCellValue(constantes.PestanaGeneral, "O2", "FACTURA AD MONTO")
	file.SetCellValue(constantes.PestanaGeneral, "P2", "FACTURA AD DESDE")
	file.SetCellValue(constantes.PestanaGeneral, "Q2", "B21 MONTO")
	file.SetCellValue(constantes.PestanaGeneral, "R2", "B21 DESDE")
	file.SetCellValue(constantes.PestanaGeneral, "S2", "COMENTARIO")
	file.SetCellValue(constantes.PestanaGeneral, "T2", "HABILITADO")
	file.SetCellValue(constantes.PestanaGeneral, "U2", "FECHA DE BAJA")
	//file.SetCellValue(constantes.PestanaGeneral, "V2", "CECOS")
	file.SetCellValue(constantes.PestanaGeneral, "V2", "TELEFONO")
	file.SetCellValue(constantes.PestanaGeneral, "W2", "EMAIL LABORAL")
	file.SetCellValue(constantes.PestanaGeneral, "X2", "EMAIL PERSONAL")
	file.SetCellValue(constantes.PestanaGeneral, "Y2", "FECHA DE NACIMIENTO")
	file.SetCellValue(constantes.PestanaGeneral, "Z2", "GENERO")
	file.SetCellValue(constantes.PestanaGeneral, "AA2", "NACIONALIDAD")
	file.SetCellValue(constantes.PestanaGeneral, "AB2", "DOMICILIO CALLE")
	file.SetCellValue(constantes.PestanaGeneral, "AC2", "DOMICILIO NUMERO")
	file.SetCellValue(constantes.PestanaGeneral, "AD2", "DOMICILIO PISO")
	file.SetCellValue(constantes.PestanaGeneral, "AE2", "DOMICILIO DEPARTAMENTO")
	file.SetCellValue(constantes.PestanaGeneral, "AF2", "DOMICILIO LOCALIDAD")
	file.SetCellValue(constantes.PestanaGeneral, "AG2", "DOMICILIO PROVINCIA")
	file.SetCellValue(constantes.PestanaGeneral, "AH2", "CECO NUMERO")
	file.SetCellValue(constantes.PestanaGeneral, "AI2", "CECO CLIENTE")
	file.SetCellValue(constantes.PestanaGeneral, "AJ2", "CECO NOMBRE")
	file.SetCellValue(constantes.PestanaGeneral, "AK2", "CECO PORCENTAJE")

	return nil
}

func initializeExcelAdmin(file *excelize.File) error {
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
	// Facturacion de servicios
	// Unir celdas de Facturacion de servicios
	file.MergeCell(constantes.PestanaFactServicios, "D1", "F1")
	file.MergeCell(constantes.PestanaFactServicios, "H1", "L1")
	// Ingresar los nombres de las celdas de Facturacion de servicios
	file.SetCellValue(constantes.PestanaFactServicios, "A2", "FECHA")
	file.SetCellValue(constantes.PestanaFactServicios, "B2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaFactServicios, "C2", "TIPO DE NOVEDAD")
	file.SetCellValue(constantes.PestanaFactServicios, "D2", "LEGAJO")
	file.SetCellValue(constantes.PestanaFactServicios, "E2", "NOMBRE")
	file.SetCellValue(constantes.PestanaFactServicios, "F2", "APELLIDO")
	file.SetCellValue(constantes.PestanaFactServicios, "G2", "CLIENTE")
	file.SetCellValue(constantes.PestanaFactServicios, "H2", "CONCEPTO")
	file.SetCellValue(constantes.PestanaFactServicios, "I2", "PERIODO")
	file.SetCellValue(constantes.PestanaFactServicios, "J2", "IMPORTE TOTAL")
	file.SetCellValue(constantes.PestanaFactServicios, "K2", "ESTADO")
	file.SetCellValue(constantes.PestanaFactServicios, "L2", "MOTIVO")
	file.SetCellValue(constantes.PestanaFactServicios, "M2", "PROVEEDOR")
	file.SetCellValue(constantes.PestanaFactServicios, "N2", "COMENTARIOS")
	file.SetCellValue(constantes.PestanaFactServicios, "D1", "USUARIO")
	file.SetCellValue(constantes.PestanaFactServicios, "H1", "CONCEPTO DE FACTURACIÓN")
	file.SetCellValue(constantes.PestanaFactServicios, "K1", "ORDEN DE COMPRA")
	file.SetCellValue(constantes.PestanaFactServicios, "L1", "PROMOVIDO")
	file.SetCellValue(constantes.PestanaFactServicios, "M1", "ENVIAR A")
	file.SetCellValue(constantes.PestanaFactServicios, "N1", "CONTACTOS")
	// Rendicion de gastos
	// Unir celdas de Rendicion de gastos
	file.MergeCell(constantes.PestanaRendGastos, "D1", "F1")
	file.MergeCell(constantes.PestanaRendGastos, "G1", "H1")
	file.MergeCell(constantes.PestanaRendGastos, "I1", "M1")
	// Ingresar los nombres de las celdas de Rendicion de gastos
	file.SetCellValue(constantes.PestanaRendGastos, "A2", "FECHA")
	file.SetCellValue(constantes.PestanaRendGastos, "B2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaRendGastos, "C2", "TIPO DE NOVEDAD")
	file.SetCellValue(constantes.PestanaRendGastos, "D2", "LEGAJO")
	file.SetCellValue(constantes.PestanaRendGastos, "E2", "NOMBRE")
	file.SetCellValue(constantes.PestanaRendGastos, "F2", "APELLIDO")
	file.SetCellValue(constantes.PestanaRendGastos, "G2", "PROVEEDOR")
	file.SetCellValue(constantes.PestanaRendGastos, "H2", "CONTACTO")
	file.SetCellValue(constantes.PestanaRendGastos, "I2", "CONCEPTO")
	file.SetCellValue(constantes.PestanaRendGastos, "J2", "PERIODO")
	file.SetCellValue(constantes.PestanaRendGastos, "K2", "IMPORTE TOTAL")
	file.SetCellValue(constantes.PestanaRendGastos, "L2", "ESTADO")
	file.SetCellValue(constantes.PestanaRendGastos, "M2", "PRIORIDAD")
	file.SetCellValue(constantes.PestanaRendGastos, "N2", "CLIENTE")
	file.SetCellValue(constantes.PestanaRendGastos, "O2", "COMENTARIOS")
	file.SetCellValue(constantes.PestanaRendGastos, "D1", "USUARIO")
	file.SetCellValue(constantes.PestanaRendGastos, "G1", "PROVEEDOR")
	file.SetCellValue(constantes.PestanaRendGastos, "I1", "CONCEPTO DE FACTURACIÓN")
	file.SetCellValue(constantes.PestanaRendGastos, "K1", "ENVIAR A")
	file.SetCellValue(constantes.PestanaRendGastos, "L1", "CONTACTO")
	file.SetCellValue(constantes.PestanaRendGastos, "M1", "RECURSO")
	// Centro de costos
	// Unir celdas de centros de costos
	file.MergeCell(constantes.PestanaNuevoCeco, "D1", "F1")
	file.MergeCell(constantes.PestanaNuevoCeco, "G1", "H1")
	file.MergeCell(constantes.PestanaNuevoCeco, "I1", "M1")
	// Ingresar los nombres de las celdas en nuevo ceco
	file.SetCellValue(constantes.PestanaNuevoCeco, "A2", "FECHA")
	file.SetCellValue(constantes.PestanaNuevoCeco, "B2", "NOVEDAD")
	file.SetCellValue(constantes.PestanaNuevoCeco, "C2", "TIPO DE NOVEDAD")
	file.SetCellValue(constantes.PestanaNuevoCeco, "D2", "LEGAJO")
	file.SetCellValue(constantes.PestanaNuevoCeco, "E2", "NOMBRE")
	file.SetCellValue(constantes.PestanaNuevoCeco, "F2", "APELLIDO")
	file.SetCellValue(constantes.PestanaNuevoCeco, "G2", "PROVEEDOR")
	file.SetCellValue(constantes.PestanaNuevoCeco, "H2", "CONTACTO")
	file.SetCellValue(constantes.PestanaNuevoCeco, "I2", "CONCEPTO")
	file.SetCellValue(constantes.PestanaNuevoCeco, "J2", "PERIODO")
	file.SetCellValue(constantes.PestanaNuevoCeco, "K2", "IMPORTE TOTAL")
	file.SetCellValue(constantes.PestanaNuevoCeco, "L2", "ESTADO")
	file.SetCellValue(constantes.PestanaNuevoCeco, "M2", "PRIORIDAD")
	file.SetCellValue(constantes.PestanaNuevoCeco, "N2", "CLIENTE")
	file.SetCellValue(constantes.PestanaNuevoCeco, "O2", "COMENTARIOS")
	file.SetCellValue(constantes.PestanaNuevoCeco, "D1", "USUARIO")
	file.SetCellValue(constantes.PestanaNuevoCeco, "G1", "PROVEEDOR")
	file.SetCellValue(constantes.PestanaNuevoCeco, "I1", "CONCEPTO DE FACTURACIÓN")

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
