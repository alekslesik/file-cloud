package main

import (
	"log"
	"github.com/alekslesik/file-cloud/internal/pkg/app"
)

func main() {
	const op = "main()"

	app, err  := app.New()
	if err != nil {
		log.Fatalf("%s > create app error: %v", op, err)
	}

	err = app.Run()
	if err != nil {
		log.Fatalf("%s > run app error: %v", op, err)
	}
}