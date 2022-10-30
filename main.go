package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/proyectoNovedades/user"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	app := fiber.New()

	dsn := "root:password@tcp(127.0.0.1:3306)/portalDeNovedades?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	user.ConnectDatabase(db)

	app.Post("/User", user.CreateUser)
	app.Get("/User/:item", user.GetUser)
	app.Put("/User/:item", user.UpdateUser)
	app.Delete("/User/:item", user.DeleteUser)
	app.Post("/Login", user.Login)

	app.Listen(":4000")
}
