package app

import (
	"database/sql"
	"os"

	"github.com/alekslesik/file-cloud/internal/app/endpoint"
	"github.com/alekslesik/file-cloud/internal/pkg/cserror"
	"github.com/alekslesik/file-cloud/internal/pkg/mailer"
	"github.com/alekslesik/file-cloud/internal/pkg/middleware"
	"github.com/alekslesik/file-cloud/internal/pkg/model"
	"github.com/alekslesik/file-cloud/internal/pkg/router"
	"github.com/alekslesik/file-cloud/internal/pkg/session"
	tmpl "github.com/alekslesik/file-cloud/internal/pkg/template"
	"github.com/alekslesik/file-cloud/pkg/config"
	"github.com/alekslesik/file-cloud/pkg/logging"
	"github.com/alekslesik/file-cloud/pkg/models"
	"github.com/rs/zerolog/log"
)

// Declare an instance of the config struct
func loadConfig() *config.Config {
	return config.New()
}

// Declare an instance of the config struct
func initLogger(level string, file *os.File) *logging.Logger {
	const op = "app.initLogger()"

	// Create a LoggerConfig based on your requirements
	config := logging.LoggerConfig{
		Level: level,
		File:  file,
	}

	// Create a LoggerFactory with the desired configuration
	factory := logging.NewLoggerFactory(config)

	// Create a logger using the factory
	logger, err := factory.CreateLogger()
	if err != nil {
		log.Err(err).Msgf("%s > create logger", op)
		return nil
	}

	return logger
}

// Declare an instance of the session struct
func initSession(cfg *config.Config) *session.Session {
	return session.New(&cfg.Session.Secret)
}

// Declare an instance of the helpers struct
// func initHelpers(logger *logging.Logger) *helpers.Helpers {
// 	return helpers.New(logger)
// }

// Declare an instance of the config struct
func initCSError() *cserror.CSError {
	return cserror.New()
}

// Data base initialization
func initDB(cfg *config.Config) (*sql.DB, error) {
	db, err := models.OpenDB(cfg.MySQL.DSN, models.MYSQL)
	if err != nil {
		return nil, err
	}

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
func initEndpoint(template tmpl.Template, logger *logging.Logger, CSError *cserror.CSError, model *model.Model, session *session.Session, mlr  *mailer.Mailer) *endpoint.Endpoint {
	return endpoint.New(&template, logger, CSError, model, *session,  mlr)
}

// Declare an instance of the config struct
func initRouter(endpoint *endpoint.Endpoint, middleware *middleware.Middleware, session *session.Session) *router.Router {
	return router.New(endpoint, middleware, session)
}

// Declare an instance of the config struct
// TODO figure out with path
func initTemplate(logger *logging.Logger) *tmpl.Template {

	appPath := os.Getenv("APP_PATH")

	return tmpl.New(logger).NewCache(appPath + "/website/content")
}

// Declare an instance of the mailer struct
func initMailer(cfg *config.Config) *mailer.Mailer {
	return mailer.New(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Sender)
}
