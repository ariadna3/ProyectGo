package user

import (
	"github.com/gofiber/fiber/v2"

)

type User struct {
	Nombre   string
	Apellido string 
	Edad     int 
}

var lista []User

func GetUser(c*fiber.Ctx) error {
	item := c.Params("item")
	for i := 0 ; len(lista) >= i; i++ {
		if item == lista[i].Nombre {

			return c.JSON(lista[i])	
		}
	}
	return c.SendString("no existe el usuario")
}
func NewUser(c*fiber.Ctx) error {

	newUser := new(User)
    if err := c.BodyParser(newUser); err != nil {
        return c.Status(503).SendString(err.Error())
    }

	lista = append(lista, *newUser)
	
    return c.Status(201).JSON(newUser)

}

