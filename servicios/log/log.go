package log

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

type Logger struct {
}

func Log(c *fiber.Ctx) error {
	logger, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	defer logger.Close()

	log.SetOutput(logger)

	for i := 0; i <= 10; i++ {
		log.Println("error linea %v", i)
	}
	return (nil)
}
