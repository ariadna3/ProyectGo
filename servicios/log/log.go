package log

import (
	"log"
	"os"
)

type Logger struct {
}

func LogPrint(logString string) error {
	logger, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	defer logger.Close()

	log.SetOutput(logger)

	for i := 0; i <= 10; i++ {
		log.Println(logString)
	}
	return (nil)
}

func LogError(logString string) error {
	logger, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	defer logger.Close()

	log.SetOutput(logger)

	for i := 0; i <= 10; i++ {
		log.Fatal(logString)
	}
	return (nil)
}
