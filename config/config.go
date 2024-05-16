package config

import (
	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
)

func InitConfig(yamlFile string, envFile string) models.Config {
	log.Debug("[CONFIG] Initializing config")
	yamlConfig := loadConfigFromYamlFile(yamlFile)
	envConfig := loadConfigFromEnv(envFile)
	mergedConfig := mergeConfigs(yamlConfig, envConfig)
	gsmConfig := loadSecretsFromGSM(mergedConfig)
	err := validateConfig(gsmConfig)

	if err != nil {
		log.Fatal("[CONFIG] Config validation failed: ", err)
	}
	log.Info("[CONFIG] Config initialized")
	return gsmConfig
}
