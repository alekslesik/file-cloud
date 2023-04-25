package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// Create a middleware chain containing our 'standard' middleware
	// which will be used for every request our application receives.
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Create a new middleware chain containing the middleware specific to
	// our dynamic application routes.
	dynamicMiddleware := alice.New(app.session.Enable, noSurf, app.authenticate)

	// New pat router with REST
	mux := pat.New()
	// Use the new dynamic middleware chain followed by the appropriate handler function.
	mux.Get("/healthcheck", dynamicMiddleware.ThenFunc(app.healthcheckHandler))
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	mux.Get("/user/logout", dynamicMiddleware.ThenFunc(app.logoutUser))
	mux.Get("/files", dynamicMiddleware.ThenFunc(app.uploadFileForm))
	mux.Post("/files", dynamicMiddleware.ThenFunc(app.uploadFile))

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	// for end-to-end testing
	// mux.Get("/ping", http.HandlerFunc(ping))

	return standardMiddleware.Then(mux)
}
