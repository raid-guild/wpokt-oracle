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
	"github.com/dan13ram/wpokt-oracle/cosmos"
	"github.com/dan13ram/wpokt-oracle/ethereum"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	logLevel := strings.ToLower(os.Getenv("LOGGER_LEVEL"))
	if logLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

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
			log.Fatal("[MAIN] Error getting absolute path for yaml file: ", err)
		}
		log.Debug("[MAIN] Yaml file: ", absYamlPath)
	}

	var absEnvPath string
	if envPath != "" {
		absEnvPath, err = filepath.Abs(envPath)
		if err != nil {
			log.Fatal("[MAIN] Error getting absolute path for env file: ", err)
		}
		log.Debug("[MAIN] Env file: ", absEnvPath)
	}

	app.InitConfig(absYamlPath, absEnvPath)
	app.InitLogger()
	app.InitDB()

	log.Debug("[MAIN] Starting server")

	services := []service.ChainServiceInterface{}
	var wg sync.WaitGroup

	healthService := app.NewHealthService(&wg)

	nodeHealth, err := healthService.GetLastHealth()
	if err != nil {
		log.Info("[MAIN] Error getting last health: ", err)
	}

	for _, ethNetwork := range app.Config.EthereumNetworks {
		chainService := ethereum.NewEthereumChainService(ethNetwork, &wg, nodeHealth)
		services = append(services, chainService)
	}

	for _, cosmosNetwork := range app.Config.CosmosNetworks {
		chainService := cosmos.NewCosmosChainService(cosmosNetwork, &wg, nodeHealth)
		services = append(services, chainService)
	}

	wg.Add(len(services) + 1)

	for _, service := range services {
		go service.Start()
	}

	go healthService.Start(services)

	log.Info("[MAIN] Server started")

	gracefulStop := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(gracefulStop, syscall.SIGINT, syscall.SIGTERM)
	go waitForExitSignals(gracefulStop, done)
	<-done

	log.Debug("[MAIN] Server stopping")

	for _, service := range services {
		service.Stop()
	}

	healthService.Stop()

	wg.Wait()

	app.DB.Disconnect()
	log.Info("[MAIN] Server stopped")
}

func waitForExitSignals(gracefulStop chan os.Signal, done chan bool) {
	sig := <-gracefulStop
	log.Debug("[MAIN] Caught signal: ", sig)
	done <- true
}
