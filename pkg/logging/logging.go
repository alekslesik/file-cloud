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
	logPath string
}

var logFile = "log.log"

// Return new zerologer
func New(level string) *Logger {

	switch level {
	case PRODUCTION:
		//logging to file
		file, err := os.OpenFile(
			logFile,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0664,
		)
		if err != nil {
			panic(err)
		}

		defer file.Close()

		z := zerolog.New(zerolog.ConsoleWriter{Out: file, TimeFormat: time.RFC3339}).
			Level(zerolog.TraceLevel).
			With().
			Stack().
			Timestamp().
			Caller().
			Logger()

		return &Logger{z, logFile}

	default:
		z := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
			Level(zerolog.TraceLevel).
			With().
			Stack().
			Timestamp().
			Caller().
			Logger()

		return &Logger{z, ""}
	}

}

func (l *Logger) SetLogFilePath(path string) {
	l.logPath = path
}
