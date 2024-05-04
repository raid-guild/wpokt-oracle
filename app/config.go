package app

import (

	// "strings"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/app/config"
	"github.com/dan13ram/wpokt-oracle/models"
)

var (
	Config models.Config
)

func InitConfig(yamlFile string, envFile string) {
	log.Debug("[CONFIG] Initializing config")
	yamlConfig := config.LoadConfigFromYamlFile(yamlFile)
	envConfig := config.LoadConfigFromEnv(envFile)
	mergedConfig := config.MergeConfigs(yamlConfig, envConfig)
	gsmConfig := config.LoadSecretsFromGSM(mergedConfig)
	err := config.ValidateConfig(gsmConfig)

	if err != nil {
		log.Fatal("[CONFIG] Config validation failed: ", err)
	}
	Config = gsmConfig
	log.Info("[CONFIG] Config initialized")
}
