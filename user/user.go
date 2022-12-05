package user

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type User struct {
	UserId   int    `json:"userId" gorm:"primaryKey"`
	Email    string `json:"email"`
	User     string `json:"user"`
	Nombre   string `json:"nombre"`
	Password string `json:"password"`
	Apellido string `json:"apellido"`
}

type Token_auth struct {
	TokenId    int    `json:"token_id" gorm:"primaryKey"`
	Token      string `json:"token"`
	UserId     int    `json:"user_id"`
	Generacion string `json:"generacion"`
	Expiracion string `json:"expiracion"`
}

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
}

var store *session.Store = session.New()
var dbUser *gorm.DB
var client *mongo.Client

var maxAge int32 = 86400 * 30 // 30 days
var isProd bool = false       // Set to true when serving over https

func ConnectMariaDb(db *gorm.DB) {
	dbUser = db
	dbUser.AutoMigrate(&User{})
	dbUser.AutoMigrate(&Token_auth{})
}

func ConnectMongoDb(clientMongo *mongo.Client) {
	client = clientMongo
}

func ShowGoogleAuthentication(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"Title": "Inicializar",
	})
}

// Novedades
// insertar novedad
func InsertNovedad(c *fiber.Ctx) error {
	novedad := new(Novedades)
	if err := c.BodyParser(novedad); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	coll := client.Database("portalDeNovedades").Collection("novedades")
	result, err := coll.InsertOne(context.TODO(), novedad)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
	return c.JSON(novedad)
}

// obtener novedad por id
func GetNovedades(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	cursor, err := coll.Find(context.TODO(), bson.M{"idSecuencial": idNumber})
	if err != nil {
		fmt.Print(err)
	}
	var novedades []Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		fmt.Print(err)
	}
	return c.JSON(novedades)
}

// obtener todas las novedades
func GetNovedadesAll(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Print(err)
	}
	var novedades []Novedades
	if err = cursor.All(context.Background(), &novedades); err != nil {
		fmt.Print(err)
	}
	return c.JSON(novedades)
}

// Borrar novedad por id
func DeleteNovedad(c *fiber.Ctx) error {
	coll := client.Database("portalDeNovedades").Collection("novedades")
	idNumber, _ := strconv.Atoi(c.Params("id"))
	result, err := coll.DeleteOne(context.TODO(), bson.M{"idSecuencial": idNumber})
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Deleted %v documents in the trainers collection", result.DeletedCount)
	return c.SendString("novedad eliminada")
}

// usuarios
// insertar usuario
func CreateUser(c *fiber.Ctx) error {
	newUser := new(User)
	if err := c.BodyParser(newUser); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	var user User
	dbUser.Where("email = ?", newUser.Email).First(&user)
	if user.User != "" {
		return c.SendString("el usuario ya existe")
	}

	if newUser.User == "" {
		newUser.User = strings.Split(newUser.Email, "@")[0]
	}
	newUser.Password = getMD5Hash(newUser.Password)
	dbUser.Create(&newUser)

	return c.Status(201).JSON(newUser)

}

// obtener usuario por user
func GetUser(c *fiber.Ctx) error {
	token := c.Cookies("token")
	if !validateToken(token) {
		return c.SendString("token invalido")
	}

	item := c.Params("item")
	var user User
	dbUser.Where("user = ?", item).First(&user)

	if user.User == "" {
		return c.SendString("no existe el usuario")
	}
	return c.JSON(user)
}

// borrar usuario por user
func DeleteUser(c *fiber.Ctx) error {
	item := c.Params("item")
	var user User
	dbUser.Where("user = ?", item).First(&user)
	if user.User == "" {
		return c.SendString("no existe el usuario")
	}
	dbUser.Where("user = ?", item).Delete(&user)
	return c.SendString("usuario eliminado")
}

// actualizar usuario por user
func UpdateUser(c *fiber.Ctx) error {
	item := c.Params("item")
	var user User
	dbUser.Where("user = ?", item).First(&user)
	if user.User == "" {
		return c.SendString("no existe el usuario")
	}
	updateUser := new(User)
	if err := c.BodyParser(updateUser); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	newUser := user
	if updateUser.Email != "" {
		newUser.Email = updateUser.Email
	}
	if updateUser.Nombre != "" {
		newUser.Nombre = updateUser.Nombre
	}
	if updateUser.Apellido != "" {
		newUser.Apellido = updateUser.Apellido
	}
	if updateUser.Password != "" {
		newUser.Password = updateUser.Password
	}
	dbUser.Model(&user).Updates(newUser)
	return c.JSON(user)
}

// login
// logeo
func Login(c *fiber.Ctx) error {
	login := new(User)
	if err := c.BodyParser(login); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	var user User
	dbUser.Where("email = ?", login.Email).First(&user)
	if user.User == "" {
		return c.SendString("no existe el usuario")
	}
	if user.Password != getMD5Hash(login.Password) {
		return c.SendString("contraseña incorrecta")
	}
	token := generateToken(32)
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
	})
	var token_auth Token_auth
	dbUser.Where("user_id = ?", user.UserId).First(&token_auth)
	if token_auth.TokenId == 0 {
		token_auth.Token = token
		token_auth.UserId = user.UserId
		token_auth.Generacion = time.Now().Format("2006-01-02 15:04:05")
		token_auth.Expiracion = time.Now().Add(time.Hour * 24).Format("2006-01-02 15:04:05")
		dbUser.Create(&token_auth)
	} else {
		token_auth.Token = token
		token_auth.Generacion = time.Now().Format("2006-01-02 15:04:05")
		token_auth.Expiracion = time.Now().Add(time.Hour * 24).Format("2006-01-02 15:04:05")
		dbUser.Model(&token_auth).Updates(token_auth)
	}
	return c.JSON(fiber.Map{
		"token": token,
		"user":  user.User,
	})
}

// token
// validar token
func validateToken(token string) bool {
	var token_auth Token_auth
	dbUser.Where("token = ?", token).First(&token_auth)
	if token_auth.TokenId == 0 {
		return false
	}
	if token_auth.Expiracion < time.Now().Format("2006-01-02 15:04:05") {
		return false
	}
	return true
}

// generar token
func generateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// md5
func getMD5Hash(message string) string {
	hash := md5.Sum([]byte(message))
	return hex.EncodeToString(hash[:])
}
