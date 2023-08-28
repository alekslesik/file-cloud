package main

import (
	"log"

	"github.com/alekslesik/file-cloud/internal/app/endpoint"
	"github.com/alekslesik/file-cloud/internal/pkg/app"
)

// Declare a string containing the application version number. Later in the book we'll
// generate this automatically at build time, but for now we'll just store the version
// number as a hard-coded global constant.
const version = "1.0.0"



func main() {
	app, err  := app.New()
	if err != nil {
		log.Fatal(err)
	}



	app.Run()
}


