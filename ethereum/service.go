package ethereum

import (
	"fmt"
	"sync"
	"time"

	"github.com/dan13ram/wpokt-oracle/ethereum/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

func NewEthereumChainService(
	config models.EthereumNetworkConfig,
	cosmosConfig models.CosmosNetworkConfig,
	mintControllerMap map[uint32][]byte,
	mnemonic string,
	wg *sync.WaitGroup,
	nodeHealth *models.Node,
) service.ChainServiceInterface {

	var chainHealth models.ChainServiceHealth
	if nodeHealth != nil {
		for _, health := range nodeHealth.Health {
			if health.Chain.ChainID == fmt.Sprintf("%d", config.ChainID) && health.Chain.ChainType == models.ChainTypeCosmos {
				chainHealth = health
				break
			}
		}
	}

	chain := util.ParseChain(config)

	var monitorRunner service.Runner
	monitorRunner = &service.EmptyRunner{}
	if config.MessageMonitor.Enabled {
		monitorRunner = NewMessageMonitor(config, mintControllerMap, chainHealth.MessageMonitor)
	}
	monitorRunnerService := service.NewRunnerService(
		"monitor",
		monitorRunner,
		config.MessageMonitor.Enabled,
		time.Duration(config.MessageMonitor.IntervalMS)*time.Millisecond,
		chain,
	)

	var signerRunner service.Runner
	signerRunner = &service.EmptyRunner{}
	if config.MessageSigner.Enabled {
		signerRunner = NewMessageSigner(mnemonic, config, cosmosConfig, mintControllerMap)
	}
	signerRunnerService := service.NewRunnerService(
		"signer",
		signerRunner,
		config.MessageSigner.Enabled,
		time.Duration(config.MessageSigner.IntervalMS)*time.Millisecond,
		chain,
	)

	relayerRunnerService := service.NewRunnerService(
		"relayer",
		&service.EmptyRunner{},
		config.MessageRelayer.Enabled,
		time.Duration(config.MessageRelayer.IntervalMS)*time.Millisecond,
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
