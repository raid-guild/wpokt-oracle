package config

import (
	"os"

	"gopkg.in/yaml.v2"

	"github.com/dan13ram/wpokt-oracle/models"
)

func loadConfigFromYamlFile(configFile string) (models.Config, error) {
	if configFile == "" {
		logger.Debug("No yaml file provided")
		return models.Config{}, nil
	}
	logger.Debugf("Loading yaml file")
	var yamlFile, err = os.ReadFile(configFile)
	if err != nil {
		return models.Config{}, err
	}
	var config models.Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return models.Config{}, err
	}
	logger.Debugf("Config loaded from yaml")
	return config, nil
}
