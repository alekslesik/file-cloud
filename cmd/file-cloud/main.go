package main

import (
	"github.com/alekslesik/file-cloud/internal/pkg/app"
	"log"
)

func main() {
	const op = "main()"

	app := app.New()
	
	err := app.Run()
	if err != nil {
		log.Fatalf("%s > run app error: %v", op, err)
	}
}
