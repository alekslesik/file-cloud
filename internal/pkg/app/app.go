package app

import (
	"flag"
	"fmt"
	"html/template"
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
	tmplCache  map[string]*template.Template
}

func New() (*Application, error) {
	const op = "app.New()"

	cfg := loadConfig()
	logger := initLogger(cfg)

	// https
	// flag.IntVar(&cfg.port, "port", 443, "API server port")
	flag.StringVar(&cfg.AppConfig.Env, "env", "development", "Environment (development|staging|production)")
	// http
	flag.IntVar(&cfg.AppConfig.Port, "port", 80, "API server port")
	cfg.MySQL.DSN = flag.String("dsn", *cfg.MySQL.DSN, "Name SQL data Source")
	flag.Parse()

	// Initialize a new session manager
	// TODO add username to session //session = session.New([]byte(*userName))
	session := initSession(cfg)
	helpers := initHelpers(logger)
	csErrors := initCSError()

	// Open DB connection pull
	db, err := initDB(helpers, cfg)
	if err != nil {
		logger.Err(err).Msgf("%s > open db", op)
		return nil, err
	}

	model := initModel(db)
	middleware := initMiddleware(session, logger, csErrors, model)
	endpoint := initEndpoint(helpers, csErrors, model, session)
	router := initRouter(endpoint, middleware, session)
	tmplCache := initTemplateCache(logger)

	// Initialization application struct
	app := &Application{
		config:     cfg,
		logger:     logger,
		endpoint:   endpoint,
		router:     router,
		middleware: middleware,
		session:    session,
		model:      model,
		tmplCache:  tmplCache,
	}

	return app, nil
}



func (a *Application) Run() error {
	const op = "app.Run()"
	var serverErr error

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.AppConfig.Port),
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

	a.logger.Info().Msgf("Server started on http://golang.fvds.ru%s/", srv.Addr)

	<-done
	a.logger.Info().Msg("server stopped")

	return serverErr
}
