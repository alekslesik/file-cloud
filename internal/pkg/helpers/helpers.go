package helpers

import (
	"html/template"

	"net/http"

	"github.com/alekslesik/file-cloud/internal/pkg/cserror"
	"github.com/alekslesik/file-cloud/pkg/logging"
	"github.com/alekslesik/file-cloud/pkg/models"
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
