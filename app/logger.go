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

	if logLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if logLevel == "info" {
		log.SetLevel(log.InfoLevel)
	} else if logLevel == "warn" {
		log.SetLevel(log.WarnLevel)
	}

	logger.Info("Logger initialized with level: ", logLevel)
}
