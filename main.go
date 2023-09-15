package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/proyectoNovedades/servicios/actividades"
	"github.com/proyectoNovedades/servicios/constantes"
	"github.com/proyectoNovedades/servicios/excel"
	"github.com/proyectoNovedades/servicios/novedades"
	"github.com/proyectoNovedades/servicios/proveedores"
	"github.com/proyectoNovedades/servicios/recursos"
	"github.com/proyectoNovedades/servicios/userGoogle"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"

	"context"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"os"
)

type Actividades struct {
	IdActividad int    `bson:"idActividad"`
	Usuario     string `bson:"usuario"`
	Fecha       string `bson:"fecha"`
	Hora        string `bson:"hora"`
	Actividad   string `bson:"actividad"`
}

type Novedades struct {
	IdSecuencial          int                 `bson:"idSecuencial"`
	Tipo                  string              `bson:"tipo"`
	Fecha                 string              `bson:"fecha"`
	Hora                  string              `bson:"hora"`
	Usuario               string              `bson:"usuario"`
	Proveedor             string              `bson:"proveedor"`
	Periodo               string              `bson:"periodo"`
	ImporteTotal          float64             `bson:"importeTotal"`
	ConceptoDeFacturacion string              `bson:"conceptoDeFacturacion"`
	Adjuntos              []string            `bson:"adjuntos"`
	Distribuciones        []Distribuciones    `bson:"distribuciones"`
	Comentarios           string              `bson:"comentarios"`
	Promovido             bool                `bson:"promovido"`
	Cliente               string              `bson:"cliente"`
	Estado                string              `bson:"estado"`
	Motivo                string              `bson:"motivo"`
	EnviarA               string              `bson:"enviarA"`
	Contacto              string              `bson:"contacto"`
	Plazo                 string              `bson:"plazo"`
	Descripcion           string              `bson:"descripcion"`
	Recursos              []RecursosNovedades `bson:"recursos"`
	Cantidad              string              `bson:"cantidad"`
	FechaDesde            string              `bson:"fechaDesde"`
	FechaHasta            string              `bson:"fechaHasta"`
	OrdenDeCompra         string              `bson:"ordenDeCompra"`
	Resumen               string              `bson:"resumen"`
	Aprobador             string              `bson:"aprobador"`
	Prioridad             string              `bson:"prioridad"`
	Reclamo               bool                `bson:"reclamo"`
	Freelance             bool                `bson:"freelance"`
}

type TipoNovedad struct {
	IdSecuencial int    `bson:"idSecuencial"`
	Tipo         string `bson:"tipo"`
	Descripcion  string `bson:"descripcion"`
}

type Cecos struct {
	IdCecos     int    `bson:"idCecos"`
	Descripcion string `bson:"descripcioncecos"`
	Cliente     string `bson:"cliente"`
	Proyecto    string `bson:"proyecto"`
	Codigo      int    `bson:"codigo"`
}

type Distribuciones struct {
	Porcentaje float64 `bson:"porcentaje"`
	Cecos      Cecos   `bson:"cecos"`
}

type RecursosNovedades struct {
	Importe     int    `bson:"importe"`
	Comentarios string `bson:"comentarios"`
	Recurso     string `bson:"recurso"`
	Periodo     string `bson:"periodo"`
	Porc        []P    `bson:"p"`
	SbActual    bool   `bson:"sbActual"`
	Retroactivo bool   `bson:"retroactivo"`
}

type P struct {
	Cc       string  `bson:"cc"`
	PorcCC   float32 `bson:"porcCC"`
	Cantidad int     `bson:"cantidad"`
}

type Proveedores struct {
	IdProveedor int    `bson:"idProveedor"`
	NumeroDoc   int    `bson:"numeroDoc"`
	RazonSocial string `bson:"razonSocial"`
}

type Recursos struct {
	IdRecurso int    `bson:"idRecurso"`
	Nombre    string `bson:"nombre"`
	Apellido  string `bson:"apellido"`
	Legajo    string `bson:"legajo"`
	Mail      string `bson:"mail"`
	Fecha     int    `bson:"date"`
}

type Files struct {
	Nombre string `bson:"nombre"`
}

var app *fiber.App

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("No se pudo cargar el archivo .env")
	}
	err = pruebaConexionEmail()
	if err != nil {
		fmt.Println("Error en el envio de mail")
	}

	goth.UseProviders(
		google.New(os.Getenv("GOOGLEKEY"), os.Getenv("GOOGLESEC"), os.Getenv("GOOGLECALLBACK")),
	)

	constantes.AllRol = os.Getenv("ALL_ROL")

	app = fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("PUERTOCORS"),
		AllowHeaders: "*",
	}))

	connectedWithMongo := createConnectionWithMongo()
	connectedWithSql := createConnectionWithMysql()

	if connectedWithMongo {

		fmt.Println("Conectado con mongo")

		userGoogle.InsertFirstUserITP(os.Getenv("USER_EMAIL_PRINCIPAL"), "Julio", "Lanzani")

		//Actividades
		app.Post("/Actividad", actividades.InsertActividad)
		app.Get("/Actividad/:id", actividades.GetActividad)
		app.Get("/Actividad", actividades.GetActividadAll)
		app.Delete("/Actividad/:id", actividades.DeleteActividad)

		//Update estado y motivo
		app.Patch("/Novedad/:id/:estado", novedades.UpdateEstadoNovedades)
		app.Patch("/Novedad/:id", novedades.UpdateMotivoNovedades)

		//Novedades
		app.Post("/Novedad", novedades.InsertNovedad)
		app.Get("/Novedad/:id", novedades.GetNovedades)
		app.Get("/Novedad/*", novedades.GetNovedadFiltro)
		app.Delete("/Novedad/:id", novedades.DeleteNovedad)
		
		//Excel
		app.Get("/Excel/Novedad/*", excel.GetExcelFile)
		//app.Get("/Excel/PagoProveedores/*", excel.GetExcelPP)

		//Workflow novedad
		app.Post("/Workflow", novedades.InsertWorkFlow)
		app.Get("/Aprobar/Novedad/:id", novedades.AprobarWorkflow)
		app.Get("/Rechazar/Novedad/:id", novedades.RechazarWorkflow)

		//obtener adjuntos novedades
		app.Get("/Archivos/Novedad/Adjuntos/:id/*", novedades.GetFiles)
		app.Put("/Archivos/Novedad/Adjuntos/:id", novedades.UpdateFileAdd)
		app.Delete("/Archivos/Novedad/Adjuntos/:id", novedades.DeleteFile)

		//Tipo Novedades
		app.Get("/TipoNovedades", novedades.GetTipoNovedad)
		//Periodos
		app.Get("/Periodos", recursos.GetFecha)

		//Centro de Costos
		app.Post("/Cecos", novedades.InsertCecos)
		app.Get("/Cecos/", novedades.GetCecosAll)
		app.Get("/Cecos/:id", novedades.GetCecos)
		app.Post("/Cecos/Package", novedades.InsertCecosPackage)

		//Proveedores
		app.Post("/Proveedor", proveedores.InsertProveedor)
		app.Get("/Proveedor/:id", proveedores.GetProveedor)
		app.Get("/Proveedor", proveedores.GetProveedorAll)
		app.Delete("/Proveedor/:id", proveedores.DeleteProveedor)
		app.Post("/Proveedor/Package", proveedores.InsertProveedoresPackage)

		//Recursos
		app.Post("/Recurso", recursos.InsertRecurso)
		app.Get("/Recurso/:id", recursos.GetRecurso)
		app.Get("/Recurso/*", recursos.GetRecursoFilter)
		app.Get("/RecursoFiltered", recursos.GetRecursoSameManager)
		app.Get("/RecursoFiltered/Cecos", recursos.GetRecursoSameCecos)
		app.Get("/RecursoFiltered/Manager", recursos.GetRecursosEmployeeOfAManager)
		app.Delete("/Recurso/:id", recursos.DeleteRecurso)
		app.Post("/Recurso/Package", recursos.InsertRecursoPackage)
		app.Patch("/Recurso/:id/*", recursos.UpdateRecurso)
		app.Put("/Recurso", recursos.PutRecurso)
		//app.Get("/HashRecurso/:id", recursos.GetRecursoHash)

		//GoogleUser
		app.Post("/user", userGoogle.InsertUserITP)
		app.Get("/user", userGoogle.GetSelfUserITP)
		app.Get("/user/:email", userGoogle.GetUserITP)
		app.Delete("user/:email", userGoogle.DeleteUserITP)
		app.Patch("/user", userGoogle.UpdateUserITP)
		app.Get("/users", userGoogle.GetUserITPAll)

	} else {
		fmt.Println("Problema al conectarse con mongo")
	}

	if connectedWithSql {
		fmt.Println("Conectado con la base sql")
	} else {
		fmt.Println("Problema al conectado con la base sql")
	}

	//----Prueba de conexion----
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("Conexion exitosa")
	})

	err = app.Listen(os.Getenv("PUERTO"))
	if err != nil {
		fmt.Println(err)
	}
}

func createConnectionWithMongo() bool {
	uri := os.Getenv("MONGOURI")
	if uri != "" {
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
		if err != nil {
			fmt.Println(err)
			return false
		}
		// Ping the primary
		if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
			fmt.Println(err)
			return false
		}
		fmt.Println("Successfully connected and pinged.")
		novedades.ConnectMongoDb(client)
		actividades.ConnectMongoDb(client)
		proveedores.ConnectMongoDb(client)
		recursos.ConnectMongoDb(client)
		userGoogle.ConnectMongoDb(client)
		excel.ConnectMongoDb(client)
		return true
	}
	return false
}

func createConnectionWithMysql() bool {
	dsn := os.Getenv("MYSQLURI")
	if dsn != "" {
		_, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			fmt.Println(err)
			return false
		}
		fmt.Println("Successfully connected to Sql.")
		return true
	}
	return false
}

func pruebaConexionEmail() error {
	// Configuración de SMTP
	smtpHost := os.Getenv("USER_HOST")
	smtpPort := os.Getenv("USER_PORT")
	smtpUsername := os.Getenv("USER_EMAIL")
	smtpPassword := os.Getenv("USER_PASSWORD")
	emailFile := os.Getenv("USER_EMAIL_FILE")

	if emailFile != "" && smtpUsername != "" {

		//lectura de archivo
		_, err := ioutil.ReadFile(emailFile)
		if err != nil {
			return err
		}

		// Mensaje de correo electrónico
		to := []string{smtpUsername}
		from := smtpUsername
		toMsg := smtpUsername
		subject := "Prueba de email exitosa"
		body := "El email pudo leerse y enviarse correctamente"

		msg := novedades.ComposeMimeMail(toMsg, from, subject, body)

		// Autenticación y envío del correo electrónico
		auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUsername, to, msg)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
