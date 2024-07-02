package config

import (
	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/common"
	"github.com/dan13ram/wpokt-oracle/models"
)

var logger *log.Entry

func init() {
	logger = log.WithFields(log.Fields{"module": "config"})
}

func InitConfig(yamlFile string, envFile string) (common.Signer, models.Config) {
	logger.Debug("Initializing config")
	yamlConfig, err := loadConfigFromYamlFile(yamlFile)
	if err != nil {
		logger.
			WithFields(log.Fields{"error": err}).
			Fatal("Error loading yaml config")
	}
	envConfig, err := loadConfigFromEnv(envFile)
	if err != nil {
		logger.
			WithFields(log.Fields{"error": err}).
			Fatal("Error loading env config")
	}
	mergedConfig := mergeConfigs(yamlConfig, envConfig)
	gsmConfig, err := loadSecretsFromGSM(mergedConfig)
	if err != nil {
		logger.
			WithFields(log.Fields{"error": err}).
			Fatal("Error loading secrets from GSM")
	}
	signer, err := validateConfig(gsmConfig)

	if err != nil {
		logger.
			WithFields(log.Fields{"error": err}).
			Fatal("Config validation failed")
	}
	logger.Info("Initialized config")
	return signer, gsmConfig
}
