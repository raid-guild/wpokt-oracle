package config

import (
	"os"
	// "strings"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/models"
	"gopkg.in/yaml.v2"
)

func LoadConfigFromYamlFile(configFile string) models.Config {
	if configFile == "" {
		log.Debug("[CONFIG] No config file provided")
		return models.Config{}
	}
	log.Debugf("[CONFIG] Loading config file %s", configFile)
	var yamlFile, err = os.ReadFile(configFile)
	if err != nil {
		log.Warnf("[CONFIG] Error reading config file %q: %s\n", configFile, err.Error())
		return models.Config{}
	}
	var config models.Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Warnf("[CONFIG] Error unmarshalling config file %q: %s\n", configFile, err.Error())
		return models.Config{}
	}
	log.Debugf("[CONFIG] Config loaded from %s", configFile)
	return config
}
