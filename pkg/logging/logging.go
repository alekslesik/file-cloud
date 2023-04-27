package logging

import (
	"github.com/alekslesik/file-cloud/pkg/config"
	"os"
	"time"

	// "github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"
)

type Logger struct {
	zerolog.Logger
}

// Return new zerologer
func GetLogger(cfg *config.Config) Logger {

	file, err := os.OpenFile(
		"myapp.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	z := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	return Logger{z}
}
