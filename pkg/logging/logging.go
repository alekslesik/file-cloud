package logging

import (
	"errors"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	DEVELOPMENT = "development"
	PRODUCTION  = "production"
)

var ErrLevelMissing error = errors.New("logging level missing")

type Logger struct {
	zerolog.Logger
}

type LoggerConfig struct {
	Level   string
	File    *os.File
}

// Create log file in specified filePath
func CreateLogFile(filePath string) (*os.File, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

type LoggerFactory struct {
	config LoggerConfig
}

func NewLoggerFactory(config LoggerConfig) *LoggerFactory {
	return &LoggerFactory{config: config}
}

// Create new logger
func (lf *LoggerFactory) CreateLogger() (*Logger, error) {
	setGlobalLogger()

	switch lf.config.Level {
	case DEVELOPMENT:
		return getDevLogger(lf.config.File), nil
	case PRODUCTION:
		return getProdLogger(lf.config.File), nil
	}

	return nil, ErrLevelMissing
}

// Log to file only
func getProdLogger(file *os.File) *Logger {
	zerolog.TimeFieldFormat = time.RFC1123
	z := zerolog.New(file).
		Level(zerolog.TraceLevel).
		With().
		Stack().
		Timestamp().
		Caller().
		Logger()

	return &Logger{z}
}

// Log to file and console
func getDevLogger(file *os.File) *Logger {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC1123}
	multi := zerolog.MultiLevelWriter(consoleWriter, file)

	z := zerolog.New(multi).
		Level(zerolog.TraceLevel).
		With().
		Stack().
		Timestamp().
		Caller().
		Logger()

	return &Logger{z}
}

// Set global logger
func setGlobalLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}