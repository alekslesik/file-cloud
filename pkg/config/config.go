package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

// Create config struct
type Config struct {
	AppConfig struct {
		LogLevel  string
		Port      int    `env:"PORT" env-default:"80"`
		Env       string `env:"ENV" env-default:"development"`
		AdminUser struct {
			Email    string `env:"ADMIN_EMAIL" env-default:"admin"`
			Password string `env:"ADMIN_PWD" env-default:"admin"`
		}
	}
	LoggerSruct struct {
		Filename   string `env:"LOG_FILENAME" env-default:"logs/log.log"`
		MaxSize    int    `env:"LOG_MAXSIZE" env-default:"100"`
		MaxBackups int    `env:"LOG_MAXBACKUP" env-default:"3"`
		MaxAge     int    `env:"LOG_MAXAGE" env-default:"24"`
		Compress   bool   `env:"LOG_COMPRESS" env-default:"true"`
	}
	MySQL struct {
		DSN string `env:"WEB_DB_DSN" env-default:"web:Todor19901///@/file_cloud?parseTime=true"`
	}

	// IsDebug       bool `env:"IS_DEBUG" env-default:"false"`
	// IsDevelopment bool `env:"IS_DEV" env-default:"true"`
}

var instance *Config
var once sync.Once

// Return instance of config
func New() *Config {
	once.Do(func() {
		instance = &Config{}

		if err := cleanenv.ReadEnv(instance); err != nil {
			helpText := "Monolith Note System"
			help, _ := cleanenv.GetDescription(instance, &helpText)
			log.Print(help)
			log.Fatal(err)
		}
	})

	return instance
}
