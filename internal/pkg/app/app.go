package app

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"net/http"
	"time"

	"github.com/alekslesik/file-cloud/internal/app/endpoint"
	"github.com/alekslesik/file-cloud/internal/pkg/middleware"
	"github.com/alekslesik/file-cloud/internal/pkg/model"
	"github.com/alekslesik/file-cloud/internal/pkg/router"
	"github.com/alekslesik/file-cloud/internal/pkg/session"
	tmpl "github.com/alekslesik/file-cloud/internal/pkg/template"
	"github.com/alekslesik/file-cloud/pkg/config"
	"github.com/alekslesik/file-cloud/pkg/logging"
)

// Declare a string containing the application version number. Later in the book we'll
// generate this automatically at build time, but for now we'll just store the version
// number as a hard-coded global constant.

type Application struct {
	config     *config.Config
	logger     *logging.Logger
	endpoint   *endpoint.Endpoint
	router     *router.Router
	middleware *middleware.Middleware
	session    *session.Session
	model      *model.Model
	template   *tmpl.Template
	dataBase   *sql.DB
}

// Create new instance of application
func New() (*Application) {
	config := loadConfig()
	// https
	// flag.IntVar(&cfg.port, "port", 443, "API server port")
	flag.StringVar(&config.App.Env, "env", "development", "Environment (development|staging|production)")
	// http
	flag.IntVar(&config.App.Port, "port", 80, "API server port")
	config.MySQL.DSN = *flag.String("dsn", config.MySQL.DSN, "Name SQL data Source")
	flag.Parse()

	logger := initLogger(config.App.Env)

	// Initialization application struct
	app := &Application{
		config:     config,
		logger:     logger,
	}

	return app
}

// Create and start server
func (a *Application) Run() error {
	const op = "app.Run()"

	dataBase, err := initDB(a.config)
	if err != nil {
		a.logger.Err(err).Msgf("%s > open db", op)
		return err
	}
	defer dataBase.Close()
	// helpers := initHelpers(logger)
	csErrors := initCSError()

	a.session = initSession(a.config)
	// TODO add username to session //session = session.New([]byte(*userName))
	a.model = initModel(dataBase)
	a.middleware = initMiddleware(a.session, a.logger, csErrors, a.model)
	a.template = initTemplate(a.logger)
	a.endpoint = initEndpoint(*a.template, a.logger, csErrors, a.model, a.session)
	a.router = initRouter(a.endpoint, a.middleware, a.session)

	var serverErr error

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.App.Port),
		Handler: a.router.Route(),
		// TLSConfig: tlsConfig,
		// Add Idle, Read and Write timeouts to the server
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			serverErr = err
			a.logger.Err(err).Msgf("%s > failed to start server", op)
		}
	}()

	a.logger.Info().Msgf("server started on http://golang.fvds.ru%s/", srv.Addr)

	<-done
	a.logger.Info().Msg("server stopped")

	return serverErr
}

// Close database
func (a *Application) Close() {
	const op = "app.Close()"

	if err := a.dataBase.Close(); err != nil {
		a.logger.Err(err).Msgf("%s > failed to close data base", op)
	}
}
