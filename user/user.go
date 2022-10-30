package user

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

type User struct {
	Email    string `json:"email"`
	User     string `json:"user"`
	Nombre   string `json:"nombre"`
	Password string `json:"password"`
	Apellido string `json:"apellido"`
}

var lista []User

func GetUser(c *fiber.Ctx) error {
	item := c.Params("item")
	for i := 0; len(lista) > i; i++ {
		if item == lista[i].User {

			return c.JSON(lista[i])
		}
	}
	return c.SendString("no existe el usuario")
}
func CreateUser(c *fiber.Ctx) error {

	newUser := new(User)
	if err := c.BodyParser(newUser); err != nil {
		return c.Status(503).SendString(err.Error())
	}
	if newUser.User == "" {
		newUser.User = strings.Split(newUser.Email, "@")[0]
	}
	lista = append(lista, *newUser)

	return c.Status(201).JSON(newUser)

}

func DeleteUser(c *fiber.Ctx) error {
	item := c.Params("item")
	for i := 0; len(lista) > i; i++ {
		if item == lista[i].User {
			lista = append(lista[:i], lista[i+1:]...)
			return c.SendString("usuario eliminado")
		}
	}
	return c.SendString("no existe el usuario")
}

func UpdateUser(c *fiber.Ctx) error {
	item := c.Params("item")
	for i := 0; len(lista) > i; i++ {
		if item == lista[i].User {
			updateUser := new(User)
			if err := c.BodyParser(updateUser); err != nil {
				return c.Status(503).SendString(err.Error())
			}
			newUser := lista[i]
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
			lista[i] = newUser
			return c.SendString("usuario actualizado")
		}
	}
	return c.SendString("no existe el usuario")
}
