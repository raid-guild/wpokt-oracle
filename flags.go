package main

import (
	"flag"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func parseFlags() (string, string) {
	var yamlPath string
	var envPath string
	flag.StringVar(&yamlPath, "yaml", "", "path to yaml file")
	flag.StringVar(&envPath, "env", "", "path to env file")
	flag.Parse()

	var absYamlPath string
	var err error
	if yamlPath != "" {
		absYamlPath, err = filepath.Abs(yamlPath)
		if err != nil {
			logger.
				WithFields(log.Fields{"error": err}).
				Fatal("Could not get absolute path for yaml file")
			panic(err)
		}
		logger.
			WithFields(log.Fields{"yaml": absYamlPath}).
			Debug("Found yaml file")
	}

	var absEnvPath string
	if envPath != "" {
		absEnvPath, err = filepath.Abs(envPath)
		if err != nil {
			logger.WithFields(log.Fields{"error": err}).Fatal("Could not get absolute path for env file")
			panic(err)
		}
		logger.WithFields(log.Fields{"env": absEnvPath}).Debug("Found env file")
	}

	return absYamlPath, absEnvPath
}
