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
	"github.com/dan13ram/wpokt-oracle/app/service"
	cfg "github.com/dan13ram/wpokt-oracle/config"
	"github.com/dan13ram/wpokt-oracle/cosmos"
	"github.com/dan13ram/wpokt-oracle/ethereum"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	logLevel := strings.ToLower(os.Getenv("LOGGER_LEVEL"))
	if logLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	logger := log.WithFields(log.Fields{"module": "main"})

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
			logger.Fatal("Error getting absolute path for yaml file: ", err)
		}
		logger.Debug("Yaml file: ", absYamlPath)
	}

	var absEnvPath string
	if envPath != "" {
		absEnvPath, err = filepath.Abs(envPath)
		if err != nil {
			logger.Fatal("Error getting absolute path for env file: ", err)
		}
		logger.Debug("Env file: ", absEnvPath)
	}

	config := cfg.InitConfig(absYamlPath, absEnvPath)
	app.InitLogger(config.Logger)
	app.InitDB(config.MongoDB)

	logger.Debug("Starting server")

	services := []service.ChainServiceInterface{}
	var wg sync.WaitGroup

	healthService := app.NewHealthService(config, &wg)

	nodeHealth, err := healthService.GetLastHealth()
	if err != nil {
		logger.Info("Error getting last health: ", err)
	}

	for _, ethNetwork := range config.EthereumNetworks {
		chainService := ethereum.NewEthereumChainService(ethNetwork, &wg, nodeHealth)
		services = append(services, chainService)
	}

	for _, cosmosNetwork := range config.CosmosNetworks {
		chainService := cosmos.NewCosmosChainService(cosmosNetwork, &wg, nodeHealth)
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
	logger := log.WithFields(log.Fields{"package": "main"})
	sig := <-gracefulStop
	logger.Debug("Caught signal: ", sig)
	done <- true
}

