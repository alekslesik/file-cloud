package app

import (
	"crypto/tls"
	"crypto/x509"
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
func New() *Application {
	return &Application{}
}

// Create and start server
func (a *Application) Run() error {
	const op = "app.Run()"

	a.config = loadConfig()

	flag.StringVar(&a.config.App.Env, "env", logging.DEVELOPMENT, "Environment (development|staging|production)")
	flag.IntVar(&a.config.App.Port, "port", 443, "API server port")
	a.config.MySQL.DSN = *flag.String("dsn", a.config.MySQL.DSN, "Name SQL data Source")
	flag.Parse()

	logFile, err := logging.CreateLogFile(a.config.Logger.LogFilePath)
	if err != nil {
		a.logger.Err(err).Msgf("%s > create file", op)
		return err
	}

	defer logFile.Close()

	a.logger = initLogger(a.config.App.Env, logFile)

	// Initialization application struct

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

	// Get root certificate from system storage
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		a.logger.Err(err).Msgf("%s > get root certificate", op)
		return err
	}

	// Set up server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.App.Port),
		Handler: a.router.Route(),
		TLSConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			RootCAs:            rootCAs,
			InsecureSkipVerify: false,
		},
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	switch srv.Addr {
	case ":80":
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				serverErr = err
				a.logger.Err(err).Msgf("%s > failed to start server", op)
			}
		}()
		a.logger.Info().Msgf("server started on http://golang.fvds.ru%s/", srv.Addr)
	case ":443":
		go func() {
			if err := srv.ListenAndServeTLS(a.config.TLS.CertPath, a.config.TLS.KeyPath); err != nil {
				serverErr = err
				a.logger.Err(err).Msgf("%s > failed to start server", op)
			}
		}()
		a.logger.Info().Msgf("server started on https://golang.fvds.ru%s/", srv.Addr)
	default:
		a.logger.Error().Msgf("%s: port not exists %s", op, srv.Addr)
	}

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
