package main

import (
	"log"
	"github.com/alekslesik/file-cloud/internal/pkg/app"
)

func main() {
	app, err  := app.New()
	if err != nil {
		log.Fatal(err)
	}

	app.Run()
}


