package app

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/signal"
	"syscall"

	"net/http"
	"time"

	"github.com/alekslesik/file-cloud/internal/app/endpoint"
	"github.com/alekslesik/file-cloud/internal/pkg/cserror"
	"github.com/alekslesik/file-cloud/internal/pkg/helpers"
	"github.com/alekslesik/file-cloud/internal/pkg/middleware"
	"github.com/alekslesik/file-cloud/internal/pkg/model"
	"github.com/alekslesik/file-cloud/internal/pkg/router"
	"github.com/alekslesik/file-cloud/internal/pkg/session"
	"github.com/alekslesik/file-cloud/internal/pkg/templates"
	"github.com/alekslesik/file-cloud/pkg/config"
	"github.com/alekslesik/file-cloud/pkg/logging"
)

// Declare a string containing the application version number. Later in the book we'll
// generate this automatically at build time, but for now we'll just store the version
// number as a hard-coded global constant.


type Application struct {
	config        *config.Config
	logger        *logging.Logger
	endpoint      endpoint.Endpoint
	router        *router.Router
	middleware    middleware.Middleware
	session       *session.Session
	model         model.Model
	templateCache map[string]*template.Template
}

func New() (*Application, error) {
	// Declare an instance of the config struct.
	cfg := config.GetConfig()
	// Declare an instance of the logger struct.
	logger := logging.GetLogger(cfg)
	// Read the value of the port and env command-line flags into the config struct. We
	// default to using the port number 4000 and the environment "development" if no
	// corresponding flags are provided.

	// https
	// flag.IntVar(&cfg.port, "port", 443, "API server port")
	flag.StringVar(&cfg.AppConfig.Env, "env", "development", "Environment (development|staging|production)")
	// http
	flag.IntVar(&cfg.AppConfig.Port, "port", 80, "API server port")
	dsn := flag.String("dsn", cfg.MySQL.DSN, "Name SQL data Source")
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret")
	flag.Parse()

	// Open DB connection pull
	db, err := openDB(*dsn)
	if err != nil {
		logger.Fatal().Err(err)
	}
	defer db.Close()

	// Initialize new cache pattern
	templateCache, err := templates.NewTemplateCache("html/")
	if err != nil {
		logger.Fatal().Err(err)
	}

	// Initialize a new session manager
	// TODO add username to session //session = session.New([]byte(*userName))
	session := session.New(secret)
	helpers := helpers.New()
	csErrors := cserror.New()
	model := model.New(db)
	middleware := middleware.New(session, &logger, csErrors, &model)
	endpoint := endpoint.New(helpers, csErrors, model, *session)
	router := router.New(endpoint, middleware, session)

	// Initialization application struct
	app := &Application{
		config:        cfg,
		logger:        &logger,
		endpoint:      endpoint,
		router:        router,
		middleware:    middleware,
		session:       session,
		model:         model,
		templateCache: templateCache,
	}

	return app, nil
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func (a *Application) Run() {
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
			a.logger.Error().Msg("failed to start server")
		}
	}()

	a.logger.Info().Msgf("Server started on http://golang.fvds.ru%s/", srv.Addr)

	<-done
	a.logger.Info().Msg("server stopped")
}
