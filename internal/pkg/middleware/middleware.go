package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alekslesik/file-cloud/internal/pkg/cserror"
	"github.com/alekslesik/file-cloud/internal/pkg/model"
	"github.com/alekslesik/file-cloud/internal/pkg/session"
	"github.com/alekslesik/file-cloud/pkg/logging"
	"github.com/alekslesik/file-cloud/pkg/models"
	"github.com/justinas/nosurf"
)

type contextKey string

var contextKeyUser = contextKey("user")

type Middleware struct {
	ses *session.Session
	log *logging.Logger
	er  *cserror.CSError
	mdl *model.Model
}

func New(ss *session.Session, lg *logging.Logger, er *cserror.CSError, md *model.Model) *Middleware {
	return &Middleware{
		ses: ss,
		log: lg,
		er:  er,
		mdl: md,
	}
}

func (m *Middleware) SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

// Create a NoSurf middleware function which uses a customized CSRF cookie with
// the Secure, Path and HttpOnly flags set.
func (m *Middleware) NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

func (m *Middleware) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.log.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.RequestURI)

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// if panic - set Connection close
				w.Header().Set("Connection", "close")
				// return 500 internal server response
				m.er.ServerError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// If the user is not authenticated, redirect them to the login page and
// return from the middleware chain so that no subsequent handlers in
// the chain are executed.
// func (app *application) requireAuthenticatedUser(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if app.authenticatedUser(r) == nil {
// 			app.session.Put(r, "flash", "Please login")
// 			http.Redirect(w, r, "/user/login", http.StatusFound)
// 			return
// 		}

// 		// Otherwise call the next handler in the chain.
// 		next.ServeHTTP(w, r)
// 	})
// }

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if a userID value exists in the session. If this *isn't
		// present* then call the next handler in the chain as normal.
		exist := m.ses.Exists(r, "userID")
		if !exist {
			next.ServeHTTP(w, r)
			return
		}

		// Fetch the details of the current user from the database.
		// If no matching record is found, remove the (invalid) userID from
		// their session and call the next handler in the chain as normal.
		user, err := m.mdl.Users.Get(m.ses.GetInt(r, "userID"))
		if err == models.ErrNoRecord {
			m.ses.Remove(r, "userID")
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			m.er.ServerError(w, err)
			return
		}

		// Otherwise, we know that the request is coming from a valid,
		// authenticated (logged in) user. We create a new copy of the
		// request with the user information added to the request context, and
		// call the next handler in the chain *using this new copy of the request*.
		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
