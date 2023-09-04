package helpers

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/alekslesik/file-cloud/internal/pkg/cserror"
	tmpl "github.com/alekslesik/file-cloud/internal/pkg/template"
	"github.com/alekslesik/file-cloud/pkg/logging"
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
	er ClientServerError
	log logging.Logger
	tmp map[string]*template.Template
}

func New(logger logging.Logger) *Helpers {
	return &Helpers{er: cserror.New()}
}

func (h *Helpers) Render(w http.ResponseWriter, r *http.Request, name string, td *tmpl.TemplateData) {
	const op = "helpers.Render()"

	// extract pattern depending "name"
	ts, ok := h.tmp[name]
	if !ok {
		h.log.Error().Msgf("%s > pattern %s not exist", op, name)
		h.er.ServerError(w, fmt.Errorf("pattern %s not exist", name))
		return
	}

	// initialize a new buffer
	buf := new(bytes.Buffer)

	// write template to the buffer, instead straight to http.ResponseWriter
	err := ts.Execute(buf, h.AddDefaultData(td, r))
	if err != nil {
		h.log.Error().Msgf("%s > template %v not executed", op, ts)
		h.er.ServerError(w, fmt.Errorf("template %v not executed", ts))
		return
	}

	// write buffer to http.ResponseWriter
	buf.WriteTo(w)
}

// Create an addDefaultData helper. This takes a pointer to a templateData
// struct, adds the current year to the CurrentYear field, and then returns
// the pointer. Again, we're not using the *http.Request parameter at the
// moment, but we will do later in the book.
func (h *Helpers) AddDefaultData(td *tmpl.TemplateData, r *http.Request) *tmpl.TemplateData {
	if td == nil {
		td = &tmpl.TemplateData{}
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
func (h *Helpers) AuthenticatedUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		return nil
	}
	return user
}

//
func (h *Helpers) OpenDB(dsn string) (*sql.DB, error) {
	const op = "helpers.OpenDB()"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		h.log.Err(err).Msgf("%s: open db", op)
		return nil, err
	}
	if err = db.Ping(); err != nil {
		h.log.Err(err).Msgf("%s: db ping", op)
		return nil, err
	}
	return db, nil
}
