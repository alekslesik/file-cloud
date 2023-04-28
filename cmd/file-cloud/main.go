package main

import (
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"html/template"

	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/alekslesik/file-cloud/pkg/config"
	"github.com/alekslesik/file-cloud/pkg/logging"
	"github.com/alekslesik/file-cloud/pkg/models"
	"github.com/alekslesik/file-cloud/pkg/models/mysql"
	"github.com/golangcollege/sessions"
)

// Declare a string containing the application version number. Later in the book we'll
// generate this automatically at build time, but for now we'll just store the version
// number as a hard-coded global constant.
const version = "1.0.0"

//go:embed *
var embedFS embed.FS

type contextKey string

var contextKeyUser = contextKey("user")

type application struct {
	config   *config.Config
	logger   *logging.Logger
	session  *sessions.Session
	UserName string
	files    interface {
		Insert(name, fileType string, size int64) (int, error)
		Get(id int) (*models.File, error)
		All() ([]*models.File, error)
	}
	templateCache map[string]*template.Template
	users         interface {
		Insert(name, email, password string) error
		Authenticate(email, password string) (int, string, error)
		Get(id int) (*models.User, error)
	}
}

func main() {
	// Declare an instance of the config struct.
	var cfg = config.GetConfig()
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

	// Initialise new cache pattern
	templateCache, err := newTemplateCache("html/")
	if err != nil {
		logger.Fatal().Err(err)
	}

	// Initialize a new session manager
	session := sessions.New([]byte(*secret))
	// TODO add username to session //session = session.New([]byte(*userName))
	session.Lifetime = 12 * time.Hour
	session.Secure = true
	session.SameSite = http.SameSiteStrictMode

	// Initialisation application struct
	app := &application{
		config:        cfg,
		logger:        &logger,
		session:       session,
		files:         &mysql.FileModel{DB: db},
		templateCache: templateCache,
		users:         &mysql.UserModel{DB: db},
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.AppConfig.Port),
		Handler: app.routes(),
		// TLSConfig: tlsConfig,
		// Add Idle, Read and Write timeouts to the server
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info().Msgf("Server started on http://golang.fvds.ru%s/", srv.Addr)

	// Use the ListenAndServeTLS() method to start the HTTPS server. We
	// pass in the paths to the TLS certificate and corresponding private key a
	// the two parameters.
	err = srv.ListenAndServe()
	// certFile := appPath + "/src/github.com/sanjas12/new_CC/tls/cert.pem"
	// keyFile := appPath + "/src/github.com/sanjas12/new_CC/tls/key.pem"

	// err = srv.ListenAndServeTLS(certFile, keyFile)
	log.Fatal().Msg(err.Error())
	logger.Fatal().Err(err)
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
