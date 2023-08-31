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
	ep *endpoint.Endpoint
	md *middleware.Middleware
	ss *session.Session
}

func New(ep *endpoint.Endpoint, md *middleware.Middleware, ss *session.Session) *Router {
	return &Router{
		ep: ep,
		md: md,
		ss: ss,
	}
}

func (r *Router) Route() http.Handler {
	// Create a middleware chain containing our 'standard' middleware
	// which will be used for every request our application receives.
	standardMiddleware := alice.New(r.md.RecoverPanic, r.md.LogRequest, r.md.SecureHeaders)

	// Create a new middleware chain containing the middleware specific to
	// our dynamic application routes.
	dynamicMiddleware := alice.New(r.ss.Enable, r.md.NoSurf, r.md.Authenticate)

	// New pat router with REST
	mux := pat.New()
	// Use the new dynamic middleware chain followed by the appropriate handler function.
	mux.Get("/healthcheck", dynamicMiddleware.ThenFunc(r.ep.HealthcheckHandler))
	mux.Get("/", dynamicMiddleware.ThenFunc(r.ep.HomeGet))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(r.ep.UserLoginGet))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(r.ep.UserLoginPost))
	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(r.ep.UserSignupGet))
	mux.Post("/user/signup", http.HandlerFunc(r.ep.UserSignupPost))
	mux.Get("/user/logout", dynamicMiddleware.ThenFunc(r.ep.UserLogoutGet))
	mux.Get("/files", dynamicMiddleware.ThenFunc(r.ep.FileUploadGet))
	mux.Post("/files", dynamicMiddleware.ThenFunc(r.ep.FileUploadPost))

	// request type '/static/css/main.css', root of embed FS is file-cloud/
	fileServer := http.FileServer(http.Dir("/website/static"))
	mux.Get("/static/", fileServer)

	// for end-to-end testing
	// mux.Get("/ping", http.HandlerFunc(ping))

	return standardMiddleware.Then(mux)
}