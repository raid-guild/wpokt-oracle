package main

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/app"
	cfg "github.com/dan13ram/wpokt-oracle/config"
	"github.com/dan13ram/wpokt-oracle/cosmos"
	"github.com/dan13ram/wpokt-oracle/ethereum"
	"github.com/dan13ram/wpokt-oracle/health"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

var logger *log.Entry

func init() {
	logFormat := strings.ToLower(os.Getenv("LOGGER_FORMAT"))
	if logFormat == "text" {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}

	logLevel := strings.ToLower(os.Getenv("LOGGER_LEVEL"))
	if logLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	logger = log.WithFields(log.Fields{"module": "main"})
}

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

func main() {
	absYamlPath, absEnvPath := parseFlags()

	config := cfg.InitConfig(absYamlPath, absEnvPath)
	app.InitLogger(config.Logger)
	app.InitDB(config.MongoDB)

	logger.Debug("Starting server")

	services := []service.ChainServiceInterface{}
	var wg sync.WaitGroup

	healthService := health.NewHealthService(config, &wg)

	var nodeHealth *models.Node
	var err error

	if config.HealthCheck.ReadLastHealth {
		nodeHealth, err = healthService.GetLastHealth()
		if err != nil {
			logger.
				WithFields(log.Fields{"error": err}).
				Warn("Could not get last health")
		}
	}

	for _, ethNetwork := range config.EthereumNetworks {
		chainService := ethereum.NewEthereumChainService(ethNetwork, &wg, nodeHealth)
		services = append(services, chainService)
	}

	for _, cosmosNetwork := range config.CosmosNetworks {
		chainService := cosmos.NewCosmosChainService(config.Mnemonic, cosmosNetwork, &wg, nodeHealth)
		services = append(services, chainService)
	}

	wg.Add(len(services) + 1)

	for _, service := range services {
		go service.Start()
	}
	go healthService.Start(services)

	logger.Info("Server started")

	gracefulStop := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(gracefulStop, syscall.SIGINT, syscall.SIGTERM)
	go waitForExitSignals(gracefulStop, done)
	<-done

	logger.Debug("Server stopping")

	for _, service := range services {
		service.Stop()
	}
	healthService.Stop()

	wg.Wait()

	app.DB.Disconnect()

	logger.Info("Server stopped")
}

func waitForExitSignals(gracefulStop chan os.Signal, done chan bool) {
	sig := <-gracefulStop
	logger.Debug("Caught signal: ", sig)
	done <- true
}
