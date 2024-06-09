package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/common"
	cfg "github.com/dan13ram/wpokt-oracle/config"
	"github.com/dan13ram/wpokt-oracle/cosmos"
	cosmosUtil "github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/db"
	"github.com/dan13ram/wpokt-oracle/ethereum"
	ethUtil "github.com/dan13ram/wpokt-oracle/ethereum/util"
	"github.com/dan13ram/wpokt-oracle/health"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

func NewMintControllerMap(config models.Config) map[uint32][]byte {
	mintControllerMap := make(map[uint32][]byte)

	for _, ethNetwork := range config.EthereumNetworks {
		mintController, err := common.BytesFromAddressHex(ethNetwork.MintControllerAddress)
		if err != nil {
			logger.
				WithError(err).WithField("chainID", ethNetwork.ChainID).
				Fatalf("Error parsing mint controller address")
			return nil
		}
		ethChain := ethUtil.ParseChain(ethNetwork)

		mintControllerMap[ethChain.ChainDomain] = mintController
	}

	cosmosNetwork := config.CosmosNetworks[0]

	mintController, err := common.AddressBytesFromBech32(cosmosNetwork.Bech32Prefix, cosmosNetwork.MultisigAddress)

	if err != nil {
		logger.
			WithError(err).WithField("chainID", cosmosNetwork.ChainID).
			Fatalf("Error parsing mint controller address")
		return nil
	}

	cosmosChain := cosmosUtil.ParseChain(cosmosNetwork)

	mintControllerMap[cosmosChain.ChainDomain] = mintController

	return mintControllerMap
}

func main() {
	absYamlPath, absEnvPath := parseFlags()

	config := cfg.InitConfig(absYamlPath, absEnvPath)

	initLogger(config.Logger)

	db.InitDB(config.MongoDB)
	defer db.DisconnectDB()

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

	if len(config.CosmosNetworks) != 1 {
		logger.Fatalf("Only one Cosmos network is supported")
	}

	cosmosNetwork := config.CosmosNetworks[0]
	mintControllerMap := NewMintControllerMap(config)

	for _, ethNetwork := range config.EthereumNetworks {
		chainService := ethereum.NewEthereumChainService(ethNetwork, cosmosNetwork, mintControllerMap, config.Mnemonic, &wg, nodeHealth)
		services = append(services, chainService)
	}

	cosmosService := cosmos.NewCosmosChainService(cosmosNetwork, mintControllerMap, config.Mnemonic, config.EthereumNetworks, &wg, nodeHealth)
	services = append(services, cosmosService)

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

	logger.Info("Server stopped")
}

func waitForExitSignals(gracefulStop chan os.Signal, done chan bool) {
	sig := <-gracefulStop
	logger.Debug("Caught signal: ", sig)
	done <- true
}
