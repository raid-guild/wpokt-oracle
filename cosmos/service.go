package cosmos

import (
	"sync"
	"time"

	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
	log "github.com/sirupsen/logrus"
)

func NewCosmosChainService(
	config models.Config,
	wg *sync.WaitGroup,
	nodeHealth *models.Node,
) service.ChainServiceInterface {

	if len(config.CosmosNetworks) != 1 {
		log.Fatal("Only one cosmos network is supported")
		return nil
	}

	cosmosConfig := config.CosmosNetworks[0]

	var chainHealth models.ChainServiceHealth
	if nodeHealth != nil {
		for _, health := range nodeHealth.Health {
			if health.Chain.ChainID == cosmosConfig.ChainID && health.Chain.ChainType == models.ChainTypeCosmos {
				chainHealth = health
				break
			}
		}
	}

	chain := util.ParseChain(cosmosConfig)

	var monitorRunner service.Runner
	monitorRunner = &service.EmptyRunner{}
	if cosmosConfig.MessageMonitor.Enabled {
		monitorRunner = NewMessageMonitor(cosmosConfig, config.EthereumNetworks, chainHealth.MessageMonitor)
	}
	monitorRunnerService := service.NewRunnerService(
		"monitor",
		monitorRunner,
		cosmosConfig.MessageMonitor.Enabled,
		time.Duration(cosmosConfig.MessageMonitor.IntervalMS)*time.Millisecond,
		chain,
	)

	var signerRunner service.Runner
	signerRunner = &service.EmptyRunner{}
	if cosmosConfig.MessageSigner.Enabled {
		signerRunner = NewMessageSigner(config.Mnemonic, cosmosConfig)
	}

	signerRunnerService := service.NewRunnerService(
		"signer",
		signerRunner,
		cosmosConfig.MessageSigner.Enabled,
		time.Duration(cosmosConfig.MessageSigner.IntervalMS)*time.Millisecond,
		chain,
	)

	var relayerRunner service.Runner
	relayerRunner = &service.EmptyRunner{}
	if cosmosConfig.MessageRelayer.Enabled {
		relayerRunner = NewMessageRelayer(cosmosConfig, chainHealth.MessageRelayer)
	}

	relayerRunnerService := service.NewRunnerService(
		"relayer",
		relayerRunner,
		cosmosConfig.MessageRelayer.Enabled,
		time.Duration(cosmosConfig.MessageRelayer.IntervalMS)*time.Millisecond,
		chain,
	)

	return service.NewChainService(
		chain,
		monitorRunnerService,
		signerRunnerService,
		relayerRunnerService,
		wg,
	)
}
