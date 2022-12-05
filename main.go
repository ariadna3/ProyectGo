package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/proyectoNovedades/user"
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

func main() {

	key := os.Getenv("GOOGLEKEY")
	sec := os.Getenv("GOOGLESEC")
	callback := os.Getenv("GOOGLECALLBACK")

	goth.UseProviders(
		google.New(key, sec, callback),
	)
	engine := html.New("./user/template", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	connectedWithMongo := createConnectionWithMongo()
	connectedWithSql := createConnectionWithMysql()

	if connectedWithMongo {
		//Novedades
		app.Post("/Novedad", user.InsertNovedad)
		app.Get("/Novedad/:id", user.GetNovedades)
		app.Delete("/Novedad/:id", user.DeleteNovedad)
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

	//Google
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
		defer func() {
			if err = client.Disconnect(context.TODO()); err != nil {
				panic(err)
			}
		}()
		// Ping the primary
		if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
			fmt.Println(err)
			return false
		}
		fmt.Println("Successfully connected and pinged.")

		user.ConnectMongoDb(client)
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
