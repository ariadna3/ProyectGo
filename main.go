package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/proyectoNovedades/user"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	gf "github.com/shareed2k/goth_fiber"
)

const (
	key = "131501611539-6th1qbgg4qg28ojgho96adlab5tgf0bf.apps.googleusercontent.com"
	sec = "AIzaSyAbh83KDi_CiEUNODtEnIMPUbNt28IKO4A"
)

func main() {
	engine := html.New("./user/template", ".html")

	goth.UseProviders(
		google.New(key, sec, "http://localhost:4000/auth/google/callback"),
	)

	app := fiber.New(fiber.Config{
		Views: engine,
	})

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
	app.Get("/", user.ShowGoogleAuthentication)
	app.Get("/auth/:provider/callback", func(ctx *fiber.Ctx) error {
		user, err := gf.CompleteUserAuth(ctx)
		if err != nil {
			return err
		}
		ctx.JSON(user)
		return nil
	})

	app.Get("/logout/:provider", func(ctx *fiber.Ctx) error {
		gf.Logout(ctx)
		ctx.Redirect("/")
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

	app.Listen(":4000")
}
