package recursos

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"github.com/proyectoNovedades/servicios/constantes"
	"github.com/proyectoNovedades/servicios/userGoogle"
)

const adminRequired = true
const adminNotRequired = false
const anyRol = ""

var client *mongo.Client

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
	userGoogle.ConnectMongoDb(client)
}

type Recursos struct {
	IdRecurso   int          `bson:"idRecurso"`
	Gerente     string       `bson:"gerente"`
	Nombre      string       `bson:"nombre"`
	Apellido    string       `bson:"apellido"`
	Legajo      int          `bson:"legajo"`
	Mail        string       `bson:"mail"`
	Fecha       time.Time    `bson:"date"`
	FechaString string       `bson:"fechaString"`
	FechaIng    string       `bson:"fechaIng"`
	Sueldo      int          `bson:"sueldo"`
	Rcc         []P          `bson:"p"`
	Vacaciones  []Vacaciones `bson:"vacaciones"`
}

type Vacaciones struct {
	CantidadComun      int `bson:"cantidadComun"`
	CantidadPatagonian int `bson:"cantidadPatagonian"`
	CantidadOtros      int `bson:"cantidadOtros"`
	Anio               int `bson:"anio"`
}

type RecursosWithID struct {
	IdObject    primitive.ObjectID `bson:"_id"`
	IdRecurso   int                `bson:"idRecurso"`
	Gerente     string             `bson:"gerente"`
	Nombre      string             `bson:"nombre"`
	Apellido    string             `bson:"apellido"`
	Legajo      int                `bson:"legajo"`
	Mail        string             `bson:"mail"`
	Fecha       time.Time          `bson:"date"`
	FechaString string             `bson:"fechaString"`
	FechaIng    string             `bson:"fechaIng"`
	Sueldo      int                `bson:"sueldo"`
	Rcc         []P                `bson:"p"`
}

type PackageOfRecursos struct {
	Paquete []Recursos
}

type P struct {
	CcNum     string  `bson:"cc"`
	CcPorc    float32 `bson:"porcCC"`
	CcNombre  string  `bson:"ccNombre"`
	CcCliente string  `bson:"ccCliente"`
}

type Cecos struct {
	IdCecos     int    `bson:"idCecos"`
	Descripcion string `bson:"descripcioncecos"`
	Cliente     string `bson:"cliente"`
	Proyecto    string `bson:"proyecto"`
	Codigo      int    `bson:"codigo"`
}

// fecha de ingreso
func GetFecha(c *fiber.Ctx) error {
	var fecha []string
	currentTime := time.Now().AddDate(0, 1, 0)
	for i := 24; i >= 0; i-- {
		anio := strconv.Itoa(currentTime.Year())
		mes := strconv.Itoa(int(currentTime.Month()))
		fecha = append(fecha, mes+"-"+anio)
		currentTime = currentTime.AddDate(0, -1, 0)
	}
	return c.Status(200).JSON(fecha)
}

// ----Recursos----
// insertar recurso
func InsertRecurso(c *fiber.Ctx) error {

	fmt.Println("Ingreso de recurso")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	//obtencion de datos
	recurso := new(Recursos)
	if err := c.BodyParser(recurso); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	fmt.Print("obtencion de datos ")
	fmt.Println(recurso)

	err, _ := elRecursoYaExisteLegajo(recurso.Legajo)
	if err != nil {
		eliminarRecurso(recurso.Mail)
	}

	//setea la fecha
	recurso.Fecha, _ = time.Parse("02/01/2006", recurso.FechaString)

	//Obtiene el ultimo Id
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "idRecurso", Value: -1}})

	cursor, _ := coll.Find(context.TODO(), filter, opts)

	var results []Recursos
	cursor.All(context.TODO(), &results)

	if len(results) == 0 {
		recurso.IdRecurso = 0
	} else {
		recurso.IdRecurso = results[0].IdRecurso + 1
	}

	//Obtiene los datos del ceco
	collCeco := client.Database(constantes.Database).Collection(constantes.CollectionCecos)

	for pos, ceco := range recurso.Rcc {
		var intVar int
		var cecoEncontrado Cecos
		if ceco.CcNum != "" {
			intVar, err = strconv.Atoi(ceco.CcNum)
			if err != nil {
				fmt.Println(err)
				return c.Status(418).SendString(err.Error())
			}
			filter := bson.D{{Key: "codigo", Value: intVar}}

			collCeco.FindOne(context.TODO(), filter).Decode(&cecoEncontrado)

		} else {
			cecoEncontrado.Cliente = ""
			cecoEncontrado.Codigo = 0
			cecoEncontrado.Descripcion = ""
			cecoEncontrado.IdCecos = 0
			cecoEncontrado.Proyecto = ""
		}
		fmt.Print("Ceco encontrado: ")
		fmt.Println(cecoEncontrado)

		recurso.Rcc[pos].CcNombre = cecoEncontrado.Cliente
		recurso.Rcc[pos].CcCliente = cecoEncontrado.Descripcion
	}

	//Ingresa el recurso
	result, err := coll.InsertOne(context.TODO(), recurso)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.Status(200).JSON(recurso)
}

// insertar paquete de recursos
func InsertRecursoPackage(c *fiber.Ctx) error {

	fmt.Println("Ingreso de paquete de recursos")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), adminRequired, anyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	//obtencion de datos
	packageRecursos := new(PackageOfRecursos)
	if err := c.BodyParser(packageRecursos); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	fmt.Print("obtencion de datos ")
	fmt.Println(packageRecursos)
	ingresarPaqueteDeRecursos(*packageRecursos)
	return c.Status(200).JSON(packageRecursos)
}

// obtener recurso por id
func GetRecurso(c *fiber.Ctx) error {
	fmt.Println("GetRecurso")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), adminNotRequired, anyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	var recurso Recursos
	err := coll.FindOne(context.TODO(), bson.D{{Key: "idRecurso", Value: idNumber}}).Decode(&recurso)
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
		return c.Status(fiber.StatusNotFound).SendString("No encontrado")
	}

	return c.Status(200).JSON(recurso)
}

// obtener los empleados de un gerente
func GetRecursosEmployeeOfAManager(c *fiber.Ctx) error {
	fmt.Println("Employees of a manager")
	// validar el token
	error, codigo, email := userGoogle.Authorization(c.Get("Authorization"), adminNotRequired, anyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	var recurso RecursosWithID
	err := coll.FindOne(context.TODO(), bson.D{{Key: "mail", Value: email}}).Decode(&recurso)
	if err != nil {
		fmt.Print(err)
		return c.Status(fiber.StatusNotFound).SendString("No encontrado")
	}
	listaSetEmpleados := agregarEmpleados(strconv.Itoa(recurso.Legajo), make(map[string]Recursos))
	var listaEmpleados []Recursos
	for _, empleado := range listaSetEmpleados {
		listaEmpleados = append(listaEmpleados, empleado)
	}
	return c.JSON(listaEmpleados)
}

// obtener todos los recursos del gerente
func GetRecursoSameManager(c *fiber.Ctx) error {

	fmt.Println("withSameManager")
	// validar el token
	error, codigo, email := userGoogle.Authorization(c.Get("Authorization"), adminNotRequired, anyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	var recurso RecursosWithID
	err := coll.FindOne(context.TODO(), bson.D{{Key: "mail", Value: email}}).Decode(&recurso)
	if err != nil {
		fmt.Print(err)
		return c.Status(fiber.StatusNotFound).SendString("No encontrado")
	}
	cursor, err := coll.Find(context.TODO(), bson.M{"gerente": strconv.Itoa(recurso.Legajo)})
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var recursos []Recursos
	if err = cursor.All(context.Background(), &recursos); err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	return c.Status(200).JSON(recursos)
}

func GetRecursoSameCecos(c *fiber.Ctx) error {
	fmt.Println("GetRecursoSameCecos")
	// validar el token
	error, codigo, email := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	var usuario Recursos
	err2 := coll.FindOne(context.TODO(), bson.M{"mail": email}).Decode(&usuario)
	if err2 != nil {
		return c.Status(200).SendString("usuario no encontrada")
	}

	coll = client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	cecos := bson.A{}
	for _, elem := range usuario.Rcc {
		cecos = append(cecos, elem.CcNum)
	}
	orTodo := bson.D{{Key: "$or", Value: cecos}}
	filter := bson.D{{Key: "p", Value: bson.D{{Key: "$elemMatch", Value: orTodo}}}}
	fmt.Println(filter)
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var recursosEncontrados []Recursos
	if err = cursor.All(context.Background(), &recursosEncontrados); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	return c.JSON(recursosEncontrados)
}

func GetRecursoFilter(c *fiber.Ctx) error {
	fmt.Println("GetRecursoFilter")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	var busqueda bson.M = bson.M{}
	if c.Query("cecos") != "" {
		cecoSearch := bson.D{{Key: "cc", Value: c.Query("cecos")}}
		busqueda["p"] = bson.D{{Key: "$elemMatch", Value: cecoSearch}}
	}

	fmt.Println(busqueda)
	cursor, err := coll.Find(context.TODO(), busqueda)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	var recursosEncontrados []Recursos
	if err = cursor.All(context.Background(), &recursosEncontrados); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	return c.JSON(recursosEncontrados)
}

func UpdateRecurso(c *fiber.Ctx) error {
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	querys := strings.Split(string(c.Request().URI().QueryString()), "&")
	var busqueda bson.D = bson.D{}
	for _, item := range querys {
		queryEncontrada := strings.Split(item, "=")
		if len(queryEncontrada) == 2 && strings.Contains(constantes.AceptarCambiosRecursos, queryEncontrada[0]) {
			busqueda = append(busqueda, bson.E{Key: queryEncontrada[0], Value: queryEncontrada[1]})
		}
	}
	fmt.Println(busqueda)
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	filter := bson.D{{Key: "idRecurso", Value: idNumber}}
	update := bson.D{{Key: "$set", Value: busqueda}}
	fmt.Println(filter)
	fmt.Println(update)
	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	return c.JSON(result)
}

// insertar recurso
func PutRecurso(c *fiber.Ctx) error {

	fmt.Println("put de recurso")
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminRequired, constantes.AnyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	//obtencion de datos
	recurso := new(Recursos)
	if err := c.BodyParser(recurso); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}
	fmt.Print("obtencion de datos ")
	fmt.Println(recurso)

	err, idRecurso := elRecursoYaExiste(recurso.Mail)
	if err == nil {
		return c.Status(fiber.StatusNotFound).SendString("Usuario no encontrado")
	}

	//setea la fecha
	recurso.Fecha, _ = time.Parse("02/01/2006", recurso.FechaString)

	//Obtiene el ultimo Id
	recurso.IdRecurso = idRecurso

	//Obtiene los datos del ceco
	collCeco := client.Database(constantes.Database).Collection(constantes.CollectionCecos)

	for pos, ceco := range recurso.Rcc {
		var intVar int
		var cecoEncontrado Cecos
		if ceco.CcNum != "" {
			intVar, err = strconv.Atoi(ceco.CcNum)
			if err != nil {
				fmt.Println(err)
				return c.Status(418).SendString(err.Error())
			}
			filter := bson.D{{Key: "codigo", Value: intVar}}

			collCeco.FindOne(context.TODO(), filter).Decode(&cecoEncontrado)

		} else {
			cecoEncontrado.Cliente = ""
			cecoEncontrado.Codigo = 0
			cecoEncontrado.Descripcion = ""
			cecoEncontrado.IdCecos = 0
			cecoEncontrado.Proyecto = ""
		}
		fmt.Print("Ceco encontrado: ")
		fmt.Println(cecoEncontrado)

		recurso.Rcc[pos].CcNombre = cecoEncontrado.Cliente
		recurso.Rcc[pos].CcCliente = cecoEncontrado.Descripcion
	}

	//Ingresa el recurso
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	eliminarRecurso(recurso.Mail)
	result, err := coll.InsertOne(context.TODO(), recurso)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.Status(200).JSON(recurso)
}

func InsertVacacionesThisYear() error {
	ActualDate := time.Now()
	decemberDate := time.Date(ActualDate.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
	year := ActualDate.Year()
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)

	var recursos []Recursos

	cursor, err := coll.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	if err = cursor.All(context.Background(), &recursos); err != nil {
		return err
	}

	for index, recurso := range recursos {
		quantityCommon := 14
		diffYears := decemberDate.Year() - recurso.Fecha.Year()
		if diffYears == 0 {
			diffDays := decemberDate.YearDay() - recurso.Fecha.YearDay()
			if diffDays < 180 {
				quantityCommon = int(diffDays / 15)
			}
		} else if diffYears >= 5 {
			quantityCommon = 21
		} else if diffYears >= 10 {
			quantityCommon = 28
		} else if diffYears >= 20 {
			quantityCommon = 35
		}
		var vacaciones Vacaciones
		vacaciones.Anio = year
		vacaciones.CantidadComun = quantityCommon
		vacaciones.CantidadPatagonian = 7
		vacaciones.CantidadOtros = 0
		// verifica que las ultimas vacaciones no sean del mismo año
		if len(recurso.Vacaciones) != 0 {
			if recurso.Vacaciones[len(recurso.Vacaciones)-1].Anio == year {
				continue
			}
		}
		recursos[index].Vacaciones = append(recursos[index].Vacaciones, vacaciones)
	}

	for _, recurso := range recursos {
		filter := bson.D{{Key: "idRecurso", Value: recurso.IdRecurso}}
		update := bson.D{{Key: "$set", Value: recurso}}
		result, err := coll.UpdateOne(context.Background(), filter, update)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Updated %v documents in the trainers collection\n", result.ModifiedCount)
	}

	return nil
}

func UpdateVacaciones(c *fiber.Ctx) error {
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminNotRequired, anyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	//obtener el id del recurso
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	//obtencion de datos
	vacacionesSolicitadas := new(Vacaciones)
	if err := c.BodyParser(vacacionesSolicitadas); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString(err.Error())
	}

	//obtener el recurso
	var recurso Recursos
	err := coll.FindOne(context.TODO(), bson.D{{Key: "idRecurso", Value: idNumber}}).Decode(&recurso)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	//verificar que no se pase de los dias
	var encontrado bool = false
	for index, vacaciones := range recurso.Vacaciones {
		if vacaciones.Anio == vacacionesSolicitadas.Anio {
			recurso.Vacaciones[index].CantidadComun = vacacionesSolicitadas.CantidadComun
			recurso.Vacaciones[index].CantidadPatagonian = vacacionesSolicitadas.CantidadPatagonian
			recurso.Vacaciones[index].CantidadOtros = vacacionesSolicitadas.CantidadOtros
			encontrado = true
			break
		}
	}
	if !encontrado {
		return c.Status(fiber.StatusNotFound).SendString("No se encontro el año solicitado")
	}

	//actualizar el recurso
	filter := bson.D{{Key: "idRecurso", Value: idNumber}}
	update := bson.D{{Key: "$set", Value: recurso}}
	result, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	return c.JSON(result)
}

// borrar recurso por id
func DeleteRecurso(c *fiber.Ctx) error {

	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminRequired, anyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idRecurso": idNumber})
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.Status(200).SendString("recurso eliminado")
}

func DeleteAllRecurso(c *fiber.Ctx) error {
	// validar el token
	error, codigo, _ := userGoogle.Authorization(c.Get("Authorization"), constantes.AdminRequired, anyRol)
	if error != nil {
		return c.Status(codigo).SendString(error.Error())
	}

	collection := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	mongoDeleteResult, err := collection.DeleteMany(context.Background(), bson.D{})
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	fmt.Printf("Deleted %v documents in the trainers collection", mongoDeleteResult.DeletedCount)
	return c.Status(200).SendString("recurso eliminado")
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func GetRecursoHash(c *fiber.Ctx) error {
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	idNumber, _ := strconv.Atoi(c.Params("id"))
	var recurso RecursosWithID
	err := coll.FindOne(context.TODO(), bson.D{{Key: "idRecurso", Value: idNumber}}).Decode(&recurso)
	fmt.Println(coll)
	if err != nil {
		fmt.Print(err)
		return c.Status(fiber.StatusNotFound).SendString("No encontrado")
	}

	hash, err := hashPassword(recurso.IdObject.Hex())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	return c.Status(200).SendString(hash)
}

func UseVacaciones(id int, vacacionesSolicitadas Vacaciones) error {
	//obtener el id del recurso
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)

	//obtener el recurso
	var recurso Recursos
	err := coll.FindOne(context.TODO(), bson.D{{Key: "idRecurso", Value: id}}).Decode(&recurso)
	if err != nil {
		return err
	}

	//verificar que no se pase de los dias
	var verificado bool = false
	for index, vacaciones := range recurso.Vacaciones {
		if vacaciones.Anio == vacacionesSolicitadas.Anio {
			verificado = (vacaciones.CantidadComun >= vacacionesSolicitadas.CantidadComun) && (vacaciones.CantidadPatagonian >= vacacionesSolicitadas.CantidadPatagonian) && (vacaciones.CantidadOtros >= vacacionesSolicitadas.CantidadOtros)
			if verificado {
				recurso.Vacaciones[index].CantidadComun -= vacacionesSolicitadas.CantidadComun
				recurso.Vacaciones[index].CantidadPatagonian -= vacacionesSolicitadas.CantidadPatagonian
				recurso.Vacaciones[index].CantidadOtros -= vacacionesSolicitadas.CantidadOtros
			}
		}
	}
	if !verificado {
		return errors.New("No se puede realizar la solicitud, año o cantidad no valida")
	}

	//actualizar el recurso
	filter := bson.D{{Key: "idRecurso", Value: id}}
	update := bson.D{{Key: "$set", Value: recurso}}
	result, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return errors.New("No se pudo actualizar el recurso")
	}
	return nil
}

func GetLastVacaciones(id int, tipo string) (Vacaciones, error) {
	//obtener el id del recurso
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	var vacaciones Vacaciones

	//obtener el recurso
	var recurso Recursos
	err := coll.FindOne(context.TODO(), bson.D{{Key: "idRecurso", Value: id}}).Decode(&recurso)
	if err != nil {
		return vacaciones, err
	}

	//obtener las vacaciones del año mas viejo
	if len(recurso.Vacaciones) == 0 {
		return vacaciones, errors.New("No se encontraron vacaciones")
	}
	vacaciones = recurso.Vacaciones[0]
	for _, vacacion := range recurso.Vacaciones {
		if vacacion.Anio < vacaciones.Anio {
			if tipo == constantes.VacacionesComunes && vacacion.CantidadComun > 0 {
				vacaciones = vacacion
			} else if tipo == constantes.VacacionesPatagonian && vacacion.CantidadPatagonian > 0 {
				vacaciones = vacacion
			} else if tipo == constantes.VacacionesOtras && vacacion.CantidadOtros > 0 {
				vacaciones = vacacion
			}
		}
	}
	return vacaciones, nil
}

func GetRecursoInterno(email string, id int, legajo int) (error, Recursos) {
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	var recurso Recursos
	if id != 0 {
		err := coll.FindOne(context.TODO(), bson.D{{Key: "idRecurso", Value: id}}).Decode(&recurso)
		if err != nil {
			return err, recurso
		}
	} else if email != "" {
		if !strings.Contains(email, "@") {
			email = email + "@itpatagonia.com"
		}
		err := coll.FindOne(context.TODO(), bson.D{{Key: "mail", Value: email}}).Decode(&recurso)
		if err != nil {
			return err, recurso
		}
	} else if legajo != 0 {
		err := coll.FindOne(context.TODO(), bson.D{{Key: "legajo", Value: legajo}}).Decode(&recurso)
		if err != nil {
			return err, recurso
		}
	} else {
		return errors.New("No se obtuvo un usuario valido"), recurso
	}

	return nil, recurso
}

func elRecursoYaExiste(email string) (error, int) {
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	filter := bson.D{{Key: "mail", Value: email}}

	cursor, _ := coll.Find(context.TODO(), filter)

	var results []Recursos
	cursor.All(context.TODO(), &results)
	if len(results) != 0 {
		return errors.New("ya existe el usuario"), results[0].IdRecurso
	}
	return nil, 0
}

func elRecursoYaExisteLegajo(legajo int) (error, int) {
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	filter := bson.D{{Key: "legajo", Value: legajo}}

	cursor, _ := coll.Find(context.TODO(), filter)

	var results []Recursos
	cursor.All(context.TODO(), &results)
	if len(results) != 0 {
		return errors.New("ya existe el usuario"), results[0].IdRecurso
	}
	return nil, 0
}

func eliminarRecurso(email string) error {
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	result, err := coll.DeleteOne(context.TODO(), bson.M{"mail": email})
	if err != nil {
		return err
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return nil
}

func eliminarRecursoLegajo(legajo int) error {
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	result, err := coll.DeleteOne(context.TODO(), bson.M{"legajo": legajo})
	if err != nil {
		return err
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return nil
}

func ingresarPaqueteDeRecursos(paqueteDeRecursos PackageOfRecursos) {
	fmt.Println("Eliminado de los recursos: ")
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)
	collCeco := client.Database(constantes.Database).Collection(constantes.CollectionCecos)

	//Obtiene el ultimo Id
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "idRecurso", Value: -1}})

	cursor, _ := coll.Find(context.TODO(), filter, opts)

	var results []Recursos
	cursor.All(context.TODO(), &results)

	var ultimoId int

	if len(results) == 0 {
		ultimoId = 0
	} else {
		ultimoId = results[0].IdRecurso + 1
	}

	//Empieza el setteo y subida
	arrayOfResources := make([]interface{}, len(paqueteDeRecursos.Paquete))
	for index, recurso := range paqueteDeRecursos.Paquete {
		err, id := elRecursoYaExiste(recurso.Mail)
		if err != nil {
			fmt.Print(recurso.Legajo)
			fmt.Print(", ")
			eliminarRecurso(recurso.Mail)
			recurso.IdRecurso = id
		} else {
			recurso.IdRecurso = ultimoId
			ultimoId = ultimoId + 1
		}
		//setea la fecha
		recurso.Fecha, _ = time.Parse("02/01/2006", recurso.FechaIng)

		//Obtiene los datos del ceco
		for pos, ceco := range recurso.Rcc {
			var cecoEncontrado Cecos
			if ceco.CcNum != "" {
				intVar, err := strconv.Atoi(ceco.CcNum)
				if err != nil {
					fmt.Println(err)
					// terminar ejecucion del recurso actual y avisar
				}
				filter := bson.D{{Key: "codigo", Value: intVar}}

				collCeco.FindOne(context.TODO(), filter).Decode(&cecoEncontrado)

			} else {
				cecoEncontrado.Cliente = ""
				cecoEncontrado.Codigo = 0
				cecoEncontrado.Descripcion = ""
				cecoEncontrado.IdCecos = 0
				cecoEncontrado.Proyecto = ""
			}
			fmt.Print("Ceco encontrado: ")
			fmt.Println(cecoEncontrado)

			recurso.Rcc[pos].CcNombre = cecoEncontrado.Cliente
			recurso.Rcc[pos].CcCliente = cecoEncontrado.Descripcion
		}
		fmt.Print("recurso final: ")
		fmt.Println(recurso)
		arrayOfResources[index] = recurso
		paqueteDeRecursos.Paquete[index] = recurso
	}
	// Ingresa el recurso
	result, err := coll.InsertMany(context.TODO(), arrayOfResources)
	if err != nil {
		// terminar ejecucion del recurso actual y avisar
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedIDs...)
}

func obtenerEmpleadosDeUnGerente(gerente string) map[string]Recursos {
	coll := client.Database(constantes.Database).Collection(constantes.CollectionRecurso)

	//Obtiene el ultimo Id
	filter := bson.D{{Key: "gerente", Value: gerente}, {Key: "gerente", Value: bson.D{{Key: "$exists", Value: 1}}}, {Key: "gerente", Value: bson.D{{Key: "$ne", Value: ""}}}}
	opts := options.Find().SetSort(bson.D{{Key: "idRecurso", Value: -1}})

	cursor, _ := coll.Find(context.TODO(), filter, opts)

	var results []Recursos
	cursor.All(context.TODO(), &results)
	var legajosEmpleados = make(map[string]Recursos)
	for _, recurso := range results {
		legajosEmpleados[strconv.Itoa(recurso.Legajo)] = recurso
	}
	return legajosEmpleados
}

func agregarEmpleados(gerente string, listaEmpleadosCompleta map[string]Recursos) map[string]Recursos {
	listaEmpleados := obtenerEmpleadosDeUnGerente(gerente)
	for key, value := range listaEmpleadosCompleta {
		listaEmpleados[key] = value
	}
	if len(listaEmpleados) == len(listaEmpleadosCompleta) {
		return listaEmpleados
	} else {
		for empleado, _ := range obtenerEmpleadosDeUnGerente(gerente) {
			nuevaListaEmpleados := agregarEmpleados(empleado, listaEmpleados)
			for nuevoEmpleadoKey, nuevoEmpleadoValue := range nuevaListaEmpleados {
				listaEmpleados[nuevoEmpleadoKey] = nuevoEmpleadoValue
			}
		}
	}
	return listaEmpleados
}
