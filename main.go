package main

import (
	"github.com/proyectoNovedades/user"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	
	app.Post("/User", user.NewUser)
	app.Get("/User/:item", user.GetUser)

	app.Listen(":4000")
}