package main

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/models"
)

var logger *log.Entry

func init() {
	logFormat := strings.ToLower(os.Getenv("LOGGER_FORMAT"))
	if logFormat == "text" {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}

	logLevel := strings.ToLower(os.Getenv("LOGGER_LEVEL"))
	if logLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	logger = log.WithFields(log.Fields{"module": "main"})
}

func initLogger(config models.LoggerConfig) {
	logLevel := strings.ToLower(config.Level)

	logger.Debug("Initializing logger")

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
		WithField("log_format", logFormat).
		WithField("log_level", logLevel).
		Info("Initialized logger")
}
