package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/proyectoNovedades/user"
)

func main() {
	app := fiber.New()

	app.Post("/User", user.CreateUser)
	app.Get("/User/:item", user.GetUser)
	app.Put("/User/:item", user.UpdateUser)
	app.Delete("/User/:item", user.DeleteUser)

	app.Listen(":4000")
}
