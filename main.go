package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/proyectoNovedades/user"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	gf "github.com/shareed2k/goth_fiber"

	"context"
	"fmt"
)

const (
	key = "131501611539-6th1qbgg4qg28ojgho96adlab5tgf0bf.apps.googleusercontent.com"
	sec = "AIzaSyAbh83KDi_CiEUNODtEnIMPUbNt28IKO4A"
	uri = "mongodb://localhost:27017/?maxPoolSize=20&w=majority"
)

func main() {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	// Ping the primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected and pinged.")

	goth.UseProviders(
		google.New(key, sec, "http://localhost:4000/auth/google/callback"),
	)
	engine := html.New("./user/template", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	dsn := "root:password@tcp(127.0.0.1:3306)/portalDeNovedades?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	user.ConnectDatabase(db, client)

	app.Post("/User", user.CreateUser)
	app.Get("/User/:item", user.GetUser)
	app.Put("/User/:item", user.UpdateUser)
	app.Delete("/User/:item", user.DeleteUser)
	app.Post("/Login", user.Login)
	app.Post("/Novedad", user.InsertNovedad)
	app.Get("/Novedad/:id", user.GetNovedades)
	app.Delete("/Novedad/:id", user.DeleteNovedad)

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

	app.Listen(":4000")
}
