package logging

import (
	"github.com/alekslesik/file-cloud/pkg/config"
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	zerolog.Logger
}

// Return new zerologer
func GetLogger(cfg *config.Config) Logger {
	z := zerolog.New(&lumberjack.Logger{
		Filename:   cfg.LoggerSruct.Filename,
		MaxSize:    cfg.LoggerSruct.MaxSize,
		MaxBackups: cfg.LoggerSruct.MaxBackups,
		MaxAge:     cfg.LoggerSruct.MaxAge,
		Compress:   cfg.LoggerSruct.Compress,
	})

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	zerolog.TimeFieldFormat = time.DateTime

	z = z.With().Caller().Time("time", time.Now()).Logger()

	return Logger{z}
}
