package main

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/proyectoNovedades/servicios/actividades"
	"github.com/proyectoNovedades/servicios/novedades"
	"github.com/proyectoNovedades/servicios/proveedores"
	"github.com/proyectoNovedades/servicios/recursos"
	"github.com/proyectoNovedades/servicios/userGoogle"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"fmt"
)

func Test_main(t *testing.T) {
	tests := []struct {
		description  string // description of the test case
		route        string // route path to test
		expectedCode int    // expected HTTP status code
	}{
		{
			description:  "get HTTP status 200",
			route:        "/",
			expectedCode: 200,
		},
		{
			description:  "get HTTP status 404, when route is not exists",
			route:        "/not_found",
			expectedCode: 404,
		},
		{
			description:  "get novedad",
			route:        "/Novedad",
			expectedCode: 200,
		},
	}
	app := fiber.New()

	connectedWithMongo := createConnectionWithMongoTest()

	if connectedWithMongo {

		fmt.Println("Conectado con mongo")

		//Actividades
		App.Post("/Actividad", actividades.InsertActividad)
		App.Get("/Actividad/:id", actividades.GetActividad)
		App.Get("/Actividad", actividades.GetActividadAll)
		App.Delete("/Actividad/:id", actividades.DeleteActividad)

		//Update estado y motivo
		App.Patch("/Novedad/:id/:estado", novedades.UpdateEstadoNovedades)
		App.Patch("/Novedad/:id", novedades.UpdateMotivoNovedades)

		//Novedades
		App.Post("/Novedad", novedades.InsertNovedad)
		App.Get("/Novedad/:id", novedades.GetNovedades)
		App.Get("/Novedad/*", novedades.GetNovedadFiltro)
		App.Get("/Novedad", novedades.GetNovedadesAll)
		App.Delete("/Novedad/:id", novedades.DeleteNovedad)

		//obtener adjuntos novedades
		App.Get("/Archivos/Novedad/Adjuntos/:id/*", novedades.GetFiles)

		//Tipo Novedades
		App.Get("/TipoNovedades", novedades.GetTipoNovedad)

		//Centro de Costos
		App.Post("/Cecos", novedades.InsertCecos)
		App.Get("/Cecos/", novedades.GetCecosAll)
		App.Get("/Cecos/:id", novedades.GetCecos)

		//Proveedores
		App.Post("/Proveedor", proveedores.InsertProveedor)
		App.Get("/Proveedor/:id", proveedores.GetProveedor)
		App.Get("/Proveedor", proveedores.GetProveedorAll)
		App.Delete("/Proveedor/:id", proveedores.DeleteProveedor)

		//Recursos
		App.Post("/Recurso", recursos.InsertRecurso)
		App.Get("/Recurso/:id", recursos.GetRecurso)
		App.Get("/Recurso", recursos.GetRecursoAll)
		App.Delete("/Recurso/:id", recursos.DeleteRecurso)

		//GoogleUser
		App.Post("/user", userGoogle.InsertUserITP)
		App.Get("/user", userGoogle.GetSelfUserITP)
		App.Get("/user/:email", userGoogle.GetUserITP)
		App.Delete("user/:email", userGoogle.DeleteUserITP)
		App.Patch("/user", userGoogle.UpdateUserITP)

	} else {
		fmt.Println("Problema al conectarse con mongo")
	}

	// Create route with GET method for test
	app.Get("/", func(c *fiber.Ctx) error {
		// Return simple string as response
		return c.SendString("Hello, World!")
	})

	// Iterate through test single test cases
	for _, test := range tests {
		// Create a new http request with the route from the test case
		req := httptest.NewRequest("GET", test.route, nil)

		// Perform the request plain with the app,
		// the second argument is a request latency
		// (set to -1 for no latency)
		resp, _ := app.Test(req, 1)

		// Verify, if the status code is as expected
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.description)
	}
}

func createConnectionWithMongoTest() bool {
	uri := "mongodb+srv://admin:2Vhk6taMChfiqWRQ@cluster0.t3d77.gcp.mongodb.net/?retryWrites=true&w=majority"
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
		return true
	}
	return false
}
