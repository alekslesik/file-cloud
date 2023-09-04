package app

import (
	"database/sql"

	"github.com/alekslesik/file-cloud/internal/app/endpoint"
	"github.com/alekslesik/file-cloud/internal/pkg/cserror"
	"github.com/alekslesik/file-cloud/internal/pkg/helpers"
	"github.com/alekslesik/file-cloud/internal/pkg/middleware"
	"github.com/alekslesik/file-cloud/internal/pkg/model"
	"github.com/alekslesik/file-cloud/internal/pkg/router"
	"github.com/alekslesik/file-cloud/internal/pkg/session"
	tmpl "github.com/alekslesik/file-cloud/internal/pkg/template"
	"github.com/alekslesik/file-cloud/pkg/config"
	"github.com/alekslesik/file-cloud/pkg/logging"
)

// Declare an instance of the config struct
func loadConfig() *config.Config {
	return config.New()
}

// Declare an instance of the config struct
func initLogger(cfg *config.Config) *logging.Logger {
	return logging.New(cfg)
}

// Declare an instance of the session struct
func initSession(cfg *config.Config) *session.Session {
	return session.New(&cfg.Session.Secret)
}

// Declare an instance of the helpers struct
func initHelpers(logger *logging.Logger) *helpers.Helpers {
	return helpers.New(logger)
}

// Declare an instance of the config struct
func initCSError() *cserror.CSError {
	return cserror.New()
}

// Data base initialization
func initDB(helpers *helpers.Helpers, cfg *config.Config) (*sql.DB, error) {
	// db, err := helpers.OpenDB(cfg.MySQL.DSN)
	db, err := helpers.OpenDB("web:Todor1990///@tcp(localhost:3306)/file_cloud")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return db, err
}

// Declare an instance of the config struct
func initModel(db *sql.DB) *model.Model {
	return model.New(db)
}

// Declare an instance of the config struct
func initMiddleware(session *session.Session, logger *logging.Logger, CSError *cserror.CSError, model *model.Model) *middleware.Middleware {
	return middleware.New(session, logger, CSError, model)
}

// Declare an instance of the config struct
func initEndpoint(template tmpl.Template, CSError *cserror.CSError, model *model.Model, session *session.Session) *endpoint.Endpoint {
	return endpoint.New(&template, CSError, model, *session)
}

// Declare an instance of the config struct
func initRouter(endpoint *endpoint.Endpoint, middleware *middleware.Middleware, session *session.Session) *router.Router {
	return router.New(endpoint, middleware, session)
}

// Declare an instance of the config struct
// TODO figure out with path
func initTemplate(logger *logging.Logger) *tmpl.Template {
	return tmpl.New(logger).NewCache("/root/go/src/github.com/alekslesik/file-cloud/website/content")
}
