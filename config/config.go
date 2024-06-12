package config

import (
	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/models"
)

var logger *log.Entry

func init() {
	logger = log.WithFields(log.Fields{"module": "config"})
}

func InitConfig(yamlFile string, envFile string) models.Config {
	logger.Debug("Initializing config")
	yamlConfig := loadConfigFromYamlFile(yamlFile)
	envConfig := loadConfigFromEnv(envFile)
	mergedConfig := mergeConfigs(yamlConfig, envConfig)
	gsmConfig := loadSecretsFromGSM(mergedConfig)
	err := validateConfig(gsmConfig)

	if err != nil {
		logger.
			WithFields(log.Fields{"error": err}).
			Fatal("Config validation failed")
	}
	logger.Info("Initialized config")
	return gsmConfig
}
