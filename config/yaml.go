package config

import (
	"os"

	"gopkg.in/yaml.v2"

	"github.com/dan13ram/wpokt-oracle/models"
)

func loadConfigFromYamlFile(configFile string) models.Config {
	if configFile == "" {
		logger.Debug("No yaml file provided")
		return models.Config{}
	}
	logger.Debugf("Loading yaml file")
	var yamlFile, err = os.ReadFile(configFile)
	if err != nil {
		logger.
			WithError(err).
			Warnf("Error reading yaml file")
		return models.Config{}
	}
	var config models.Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		logger.
			WithError(err).
			Warnf("Error unmarshalling yaml file")
		return models.Config{}
	}
	logger.Debugf("Config loaded from yaml")
	return config
}
