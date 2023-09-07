package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Port      int    `env:"PORT" env-default:"443"`
	Env       string `env:"ENV" env-default:"development"`
	AdminUser struct {
		Email    string `env:"ADMIN_EMAIL" env-default:"admin"`
		Password string `env:"ADMIN_PWD" env-default:"admin"`
	}
}

type LoggerConfig struct {
	LogFilePath string `env:"LOG_FILE" env-default:"tmp/log.log"`
	MaxSize     int    `env:"LOG_MAXSIZE" env-default:"100"`
	MaxBackups  int    `env:"LOG_MAXBACKUP" env-default:"3"`
	MaxAge      int    `env:"LOG_MAXAGE" env-default:"24"`
	Compress    bool   `env:"LOG_COMPRESS" env-default:"true"`
}

type MySQLConfig struct {
	DSN string `env:"WEB_DB_DSN" env-default:"web:Todor1990///@tcp(localhost:3306)/file_cloud?parseTime=true"`
}

type TlsConfig struct {
	KeyPath  string `env:"TLS_KEY_PATH" env-default:"./tls/key.pem"`
	CertPath string `env:"TLS_CERT_PATH" env-default:"./tls/cert.pem"`
}

type SessionConfig struct {
	Secret string `env:"SESSION_SECRET" env-default:"s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge"`
}

type Config struct {
	App     AppConfig
	Logger  LoggerConfig
	MySQL   MySQLConfig
	Session SessionConfig
	TLS     TlsConfig
}

// Singleton pattern
var instance *Config
var once sync.Once

// Return instance of config
func New() *Config {
	once.Do(func() {
		instance = &Config{}

		if err := cleanenv.ReadEnv(instance); err != nil {
			helpText := "File cloud"
			help, _ := cleanenv.GetDescription(instance, &helpText)
			log.Print(help)
			log.Fatal(err)
		}
	})

	return instance
}
