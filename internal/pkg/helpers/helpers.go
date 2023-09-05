package helpers

import (
	// "bytes"
	"database/sql"
	// "fmt"
	"html/template"

	"net/http"
	// "time"

	"github.com/alekslesik/file-cloud/internal/pkg/cserror"
	// tmpl "github.com/alekslesik/file-cloud/internal/pkg/template"
	"github.com/alekslesik/file-cloud/pkg/logging"
	"github.com/alekslesik/file-cloud/pkg/models"
	// "github.com/justinas/nosurf"
)

type contextKey string

var contextKeyUser = contextKey("user")

type ClientServerError interface {
	ClientError(http.ResponseWriter, int, error)
	ServerError(http.ResponseWriter, error)
}

type tmplCache map[string]*template.Template

type Helpers struct {
	er  ClientServerError
	log *logging.Logger
	tmp map[string]*template.Template
}

func New(logger *logging.Logger) *Helpers {
	return &Helpers{
		er:  cserror.New(),
		log: logger,
		tmp: make(tmplCache),
	}
}

// Return userID ID from session
func AuthenticatedUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		return nil
	}
	return user
}

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



