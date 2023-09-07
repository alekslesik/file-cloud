package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

const (
	DEVELOPMENT = "development"
	PRODUCTION  = "production"
)

type Logger struct {
	zerolog.Logger
}

// Return new zerologer
func New(level string, file *os.File) *Logger {
	switch level {
	case PRODUCTION:
		z := zerolog.New(zerolog.ConsoleWriter{Out: file, TimeFormat: time.RFC3339}).
			Level(zerolog.TraceLevel).
			With().
			Stack().
			Timestamp().
			Caller().
			Logger()

		return &Logger{z}
	default:
		z := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
			Level(zerolog.TraceLevel).
			With().
			Stack().
			Timestamp().
			Caller().
			Logger()

		return &Logger{z,}
	}

}