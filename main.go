package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html"
	"github.com/proyectoNovedades/servicios/actividades"
	"github.com/proyectoNovedades/servicios/novedades"
	"github.com/proyectoNovedades/servicios/proveedores"
	"github.com/proyectoNovedades/servicios/recursos"
	"github.com/proyectoNovedades/servicios/user"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	gf "github.com/shareed2k/goth_fiber"

	"context"
	"fmt"
	"os"
)

type Cecos struct {
	IdCeco      int    `bson:"id_ceco"`
	Descripcion string `bson:"descripcion"`
}

type Distribuciones struct {
	Porcentaje float64 `bson:"porcentaje"`
	Ceco       Cecos   `bson:"ceco"`
}

type Novedades struct {
	IdSecuencial          int              `bson:"idSecuencial"`
	Tipo                  string           `bson:"tipo"`
	Descripcion           string           `bson:"descripcion"`
	Fecha                 string           `bson:"fecha"`
	Hora                  string           `bson:"hora"`
	Usuario               string           `bson:"usuario"`
	Proveedor             string           `bson:"proveedor"`
	Periodo               string           `bson:"periodo"`
	ImporteTotal          float64          `bson:"importeTotal"`
	ConceptoDeFacturacion string           `bson:"conceptoDeFacturacion"`
	Adjuntos              []string         `bson:"adjuntos"`
	Distribuciones        []Distribuciones `bson:"distribuciones"`
	Comentarios           string           `bson:"comentarios"`
	Promovido             bool             `bson:"promovido"`
	Cliente               string           `bson:"cliente"`
}

type TipoNovedad struct {
	IdSecuencial int    `bson:"idSecuencial"`
	Tipo         string `bson:"tipo"`
	Descripcion  string `bson:"descripcion"`
}

type Actividades struct {
	IdNovedad int    `bson:"idNovedad"`
	Usuario   string `bson:"usuario"`
	Fecha     string `bson:"fecha"`
	Hora      string `bson:"hora"`
	Actividad string `bson:"actividad"`
}

type Proveedores struct {
	IdProveedor int    `bson:"idProveedor"`
	NumeroDoc   int    `bson:"numeroDoc"`
	RazonSocial string `bson:"razonSocial"`
}

func main() {

	goth.UseProviders(
		google.New(os.Getenv("GOOGLEKEY"), os.Getenv("GOOGLESEC"), os.Getenv("GOOGLECALLBACK")),
	)
	engine := html.New("./servicios/user/template", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("PUERTOCORS"),
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	connectedWithMongo := createConnectionWithMongo()
	connectedWithSql := createConnectionWithMysql()

	if connectedWithMongo {
		//Update estado y motivo
		app.Patch("/Novedad/:id/:estado", novedades.UpdateEstadoNovedades)
		app.Patch("/Novedad/:id", novedades.UpdateMotivoNovedades)

		//Novedades
		app.Post("/Novedad", novedades.InsertNovedad)
		app.Get("/Novedad/:id", novedades.GetNovedades)
		app.Get("/Novedad/*", novedades.GetNovedadFiltro)
		app.Get("/Novedad", novedades.GetNovedadesAll)
		app.Delete("/Novedad/:id", novedades.DeleteNovedad)
		app.Patch("/Novedad/:id/:estado", novedades.UpdateEstadoNovedades)

		//Tipo Novedades
		app.Get("/TipoNovedades", novedades.GetTipoNovedad)

		//Actividades
		app.Post("/Actividad", actividades.InsertActividad)
		app.Get("/Actividad/:id", actividades.GetActividad)
		app.Get("/Actividad", actividades.GetActividadAll)
		app.Delete("/Actividad/:id", actividades.DeleteActividad)

		//Proveedores
		app.Post("/Proveedor", proveedores.InsertProveedor)
		app.Get("/Proveedor/:id", proveedores.GetProveedor)
		app.Get("/Proveedor", proveedores.GetProveedorAll)
		app.Delete("/Proveedor/:id", proveedores.DeleteProveedor)

		//Centro de Costos
		app.Get("/Cecos/", novedades.GetCecosAll)
		app.Get("/Cecos/:id", novedades.GetCecos)

		//Recursos
		app.Post("/Recurso", recursos.InsertRecurso)
		app.Get("/Recurso/:id", recursos.GetRecurso)
		app.Get("/Recurso", recursos.GetRecursoAll)
		app.Delete("/Recurso/:id", recursos.DeleteRecurso)

	}

	if connectedWithSql {
		//User
		app.Post("/User", user.CreateUser)
		app.Get("/User/:item", user.GetUser)
		app.Put("/User/:item", user.UpdateUser)
		app.Delete("/User/:item", user.DeleteUser)

		//Login
		app.Post("/Login", user.Login)
	}

	//----Google----
	app.Get("/", user.ShowGoogleAuthentication)

	app.Get("/auth/:provider/callback", func(ctx *fiber.Ctx) error {
		user, err := gf.CompleteUserAuth(ctx)
		if err != nil {
			return err
		}
		ctx.JSON(user)
		return nil
	})

	app.Get("/auth/:provider", func(ctx *fiber.Ctx) error {
		if gothUser, err := gf.CompleteUserAuth(ctx); err == nil {
			ctx.JSON(gothUser)
		} else {
			gf.BeginAuthHandler(ctx)
		}
		return nil
	})

	app.Get("/logout/:provider", func(ctx *fiber.Ctx) error {
		gf.Logout(ctx)
		ctx.Redirect("/")
		return nil
	})

	fmt.Println(os.Getenv("PUERTO"))
	app.Listen(os.Getenv("PUERTO"))
}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		panic("Error loading .env file")
	}

	return os.Getenv(key)
}

func createConnectionWithMongo() bool {
	uri := goDotEnvVariable("MONGOURI")
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
		user.ConnectMongoDb(client)
		novedades.ConnectMongoDb(client)
		actividades.ConnectMongoDb(client)
		proveedores.ConnectMongoDb(client)
		recursos.ConnectMongoDb(client)
		return true
	}
	return false
}

func createConnectionWithMysql() bool {
	dsn := goDotEnvVariable("MYSQLURI")
	if dsn != "" {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			fmt.Println(err)
			return false
		}
		fmt.Println("Successfully connected to Sql.")
		user.ConnectMariaDb(db)
		return true
	}
	return false
}
