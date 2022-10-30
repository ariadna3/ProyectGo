package user

import (
	"strings"
	"time"

	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
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

var store *session.Store = session.New()
var lista []User
var dbUser *gorm.DB

func ConnectDatabase(db *gorm.DB) {
	dbUser = db
	println(dbUser)
	println(db)
	dbUser.AutoMigrate(&User{})
	dbUser.AutoMigrate(&Token_auth{})
}

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

func generateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func getMD5Hash(message string) string {
	hash := md5.Sum([]byte(message))
	return hex.EncodeToString(hash[:])
}
