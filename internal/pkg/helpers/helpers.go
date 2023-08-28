package helpers

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/alekslesik/file-cloud/internal/pkg/cserror"
	"github.com/alekslesik/file-cloud/internal/pkg/templates"
	"github.com/alekslesik/file-cloud/pkg/models"
	"github.com/justinas/nosurf"
)

type contextKey string

var contextKeyUser = contextKey("user")

type ClientServerError interface {
	ClientError(http.ResponseWriter, int, error)
	ServerError(http.ResponseWriter, error)
}

type Helpers struct {
	e ClientServerError
	templateCache map[string]*template.Template
}

func New() Helpers {
	return Helpers{e: cserror.New()}
}

func (h Helpers) Render(w http.ResponseWriter, r *http.Request, name string, td *templates.TemplateData) {
	// extract pattern depending "name"
	ts, ok := h.templateCache[name]
	if !ok {
		h.e.ServerError(w, fmt.Errorf("pattern %s not exist", name))
		return
	}

	// initialize a new buffer
	buf := new(bytes.Buffer)

	// write template to the buffer, instead straight to http.ResponseWriter
	err := ts.Execute(buf, h.AddDefaultData(td, r))
	if err != nil {
		h.e.ServerError(w, fmt.Errorf("template %v not executed", ts))
		return
	}

	// write buffer to http.ResponseWriter
	buf.WriteTo(w)
}

// Create an addDefaultData helper. This takes a pointer to a templateData
// struct, adds the current year to the CurrentYear field, and then returns
// the pointer. Again, we're not using the *http.Request parameter at the
// moment, but we will do later in the book.
func (h Helpers) AddDefaultData(td *templates.TemplateData, r *http.Request) *templates.TemplateData {
	if td == nil {
		td = &templates.TemplateData{}
	}

	// Add current time.
	td.CurrentYear = time.Now().Year()
	// Add flash message.
	// TODO sort out
	// td.Flash = app.session.PopString(r, "flash")
	// Check if user is authenticate.
	td.AuthenticatedUser = h.AuthenticatedUser(r)
	// Add the CSRF token to the templateData struct.
	td.CSRFToken = nosurf.Token(r)
	// Add User Name to template
	// td.UserName = app.UserName

	return td
}

// Return userID ID from session
func (h Helpers) AuthenticatedUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		return nil
	}
	return user
}
