package cosmos

import (
	"sync"
	"time"

	"github.com/dan13ram/wpokt-oracle/cosmos/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

func NewCosmosChainService(
	config models.CosmosNetworkConfig,
	mintControllerMap map[uint32][]byte,
	mnemonic string,
	ethNetworks []models.EthereumNetworkConfig,
	wg *sync.WaitGroup,
	nodeHealth *models.Node,
) service.ChainService {

	var chainHealth models.ChainServiceHealth
	if nodeHealth != nil {
		for _, health := range nodeHealth.Health {
			if health.Chain.ChainID == config.ChainID && health.Chain.ChainType == models.ChainTypeCosmos {
				chainHealth = health
				break
			}
		}
	}

	chain := util.ParseChain(config)

	var monitorRunnable service.Runnable = &service.EmptyRunnable{}
	if config.MessageMonitor.Enabled {
		monitorRunnable = NewMessageMonitor(config, mintControllerMap, ethNetworks, chainHealth.MessageMonitor)
	}
	monitorRunnerService := service.NewRunnerService(
		"monitor",
		monitorRunnable,
		config.MessageMonitor.Enabled,
		time.Duration(config.MessageMonitor.IntervalMS)*time.Millisecond,
		chain,
	)

	var signerRunnable service.Runnable = &service.EmptyRunnable{}
	if config.MessageSigner.Enabled {
		signerRunnable = NewMessageSigner(mnemonic, config, mintControllerMap, ethNetworks)
	}

	signerRunnerService := service.NewRunnerService(
		"signer",
		signerRunnable,
		config.MessageSigner.Enabled,
		time.Duration(config.MessageSigner.IntervalMS)*time.Millisecond,
		chain,
	)

	var relayerRunnable service.Runnable = &service.EmptyRunnable{}
	if config.MessageRelayer.Enabled {
		relayerRunnable = NewMessageRelayer(config, chainHealth.MessageRelayer)
	}

	relayerRunnerService := service.NewRunnerService(
		"relayer",
		relayerRunnable,
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
