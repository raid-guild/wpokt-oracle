package app

import (
	"strings"

	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
)

func InitLogger(config models.LoggerConfig) {
	logLevel := strings.ToLower(config.Level)
	logger := log.WithField("module", "logger")
	logger.Debug("Initializing logger with level: ", logLevel)

	switch logLevel {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	logFormat := strings.ToLower(config.Format)

	switch logFormat {
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.SetFormatter(&log.JSONFormatter{})
	}

	logger.
		WithField("format", logFormat).
		Info("Logger initialized")
}
