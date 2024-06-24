package main

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dan13ram/wpokt-oracle/models"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		config         models.LoggerConfig
		expectedLevel  log.Level
		expectedFormat log.Formatter
	}{
		{models.LoggerConfig{Level: "debug", Format: "text"}, log.DebugLevel, &log.TextFormatter{}},
		{models.LoggerConfig{Level: "info", Format: "json"}, log.InfoLevel, &log.JSONFormatter{}},
		{models.LoggerConfig{Level: "warn", Format: "json"}, log.WarnLevel, &log.JSONFormatter{}},
		{models.LoggerConfig{Level: "error", Format: "text"}, log.ErrorLevel, &log.TextFormatter{}},
		{models.LoggerConfig{Level: "fatal", Format: "json"}, log.FatalLevel, &log.JSONFormatter{}},
		{models.LoggerConfig{Level: "panic", Format: "text"}, log.PanicLevel, &log.TextFormatter{}},
		{models.LoggerConfig{Level: "trace", Format: "json"}, log.TraceLevel, &log.JSONFormatter{}},
		{models.LoggerConfig{Level: "unknown", Format: "unknown"}, log.InfoLevel, &log.JSONFormatter{}},
	}

	for _, test := range tests {
		t.Run(test.config.Level+"_"+test.config.Format, func(t *testing.T) {
			initLogger(test.config)

			assert.Equal(t, test.expectedLevel, log.GetLevel())

			formatter := log.StandardLogger().Formatter
			require.IsType(t, test.expectedFormat, formatter)
		})
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		envLevel       string
		envFormat      string
		expectedLevel  log.Level
		expectedFormat log.Formatter
	}{
		{"debug", "text", log.DebugLevel, &log.TextFormatter{}},
		{"info", "json", log.InfoLevel, &log.JSONFormatter{}},
		{"unknown", "unknown", log.InfoLevel, &log.JSONFormatter{}},
	}

	for _, test := range tests {
		t.Run(test.envLevel+"_"+test.envFormat, func(t *testing.T) {
			os.Setenv("LOGGER_LEVEL", test.envLevel)
			os.Setenv("LOGGER_FORMAT", test.envFormat)
			loggerInit()

			assert.Equal(t, test.expectedLevel, log.GetLevel())

			formatter := log.StandardLogger().Formatter
			require.IsType(t, test.expectedFormat, formatter)

			// Clean up environment variables
			os.Unsetenv("LOGGER_LEVEL")
			os.Unsetenv("LOGGER_FORMAT")
		})
	}
}
