package app

import (
	"strings"

	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
)

func InitLogger(config models.LoggerConfig) {
	logLevel := strings.ToLower(config.Level)
	log.Debug("[LOGGER] Initializing logger with level: ", logLevel)

	if logLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if logLevel == "info" {
		log.SetLevel(log.InfoLevel)
	} else if logLevel == "warn" {
		log.SetLevel(log.WarnLevel)
	}

	log.Info("[LOGGER] Logger initialized with level: ", logLevel)
}
