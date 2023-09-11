package router

import (
	"net/http"

	"github.com/alekslesik/file-cloud/internal/app/endpoint"
	"github.com/alekslesik/file-cloud/internal/pkg/middleware"
	"github.com/alekslesik/file-cloud/internal/pkg/session"
	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

type Router struct {
	edp *endpoint.Endpoint
	mdw *middleware.Middleware
	ses *session.Session
}

func New(edp *endpoint.Endpoint, mdw *middleware.Middleware, ses *session.Session) *Router {
	return &Router{
		edp: edp,
		mdw: mdw,
		ses: ses,
	}
}

func (r *Router) Route() http.Handler {
	// Create a middleware chain containing our 'standard' middleware
	// which will be used for every request our application receives.
	standardMiddleware := alice.New(r.mdw.RecoverPanic, r.mdw.LogRequest, r.mdw.SecureHeaders)

	

	// Create a new middleware chain containing the middleware specific to
	// our dynamic application routes.
	dynamicMiddleware := alice.New(r.ses.Enable, r.mdw.NoSurf, r.mdw.Authenticate)

	// New pat router with REST
	mux := pat.New()
	// Use the new dynamic middleware chain followed by the appropriate handler function.
	mux.Get("/healthcheck", dynamicMiddleware.ThenFunc(r.edp.HealthcheckHandler))
	mux.Get("/", dynamicMiddleware.ThenFunc(r.edp.HomeGet))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(r.edp.UserLoginGet))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(r.edp.UserLoginPost))
	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(r.edp.UserSignupGet))
	// mux.Post("/user/signup", http.HandlerFunc(r.edp.UserSignupPost))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(r.edp.UserSignupPost))
	mux.Get("/user/logout", dynamicMiddleware.ThenFunc(r.edp.UserLogoutGet))
	mux.Get("/files", dynamicMiddleware.ThenFunc(r.edp.FileUploadGet))
	mux.Post("/files", dynamicMiddleware.ThenFunc(r.edp.FileUploadPost))

	// file server for static files
	fileServer := http.FileServer(http.Dir("./website/static/"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	// for end-to-end testing
	// mux.Get("/ping", http.HandlerFunc(ping))

	return standardMiddleware.Then(mux)
}
