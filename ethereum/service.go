package ethereum

import (
	"fmt"
	"sync"
	"time"

	eth "github.com/dan13ram/wpokt-oracle/ethereum/client"
	"github.com/dan13ram/wpokt-oracle/ethereum/util"
	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

var utilParseChain = util.ParseChain
var utilSignMessage = util.SignMessage
var ethValidateTransactionByHash = eth.ValidateTransactionByHash

func NewEthereumChainService(
	config models.EthereumNetworkConfig,
	cosmosConfig models.CosmosNetworkConfig,
	mintControllerMap map[uint32][]byte,
	ethNetworks []models.EthereumNetworkConfig,
	mnemonic string,
	wg *sync.WaitGroup,
	nodeHealth *models.Node,
) service.ChainService {

	var chainHealth models.ChainServiceHealth
	if nodeHealth != nil {
		for _, health := range nodeHealth.Health {
			if health.Chain.ChainID == fmt.Sprintf("%d", config.ChainID) && health.Chain.ChainType == models.ChainTypeCosmos {
				chainHealth = health
				break
			}
		}
	}

	chain := utilParseChain(config)

	var monitorRunnable service.Runnable = &service.EmptyRunnable{}
	if config.MessageMonitor.Enabled {
		monitorRunnable = NewMessageMonitor(config, mintControllerMap, chainHealth.MessageMonitor)
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
		signerRunnable = NewMessageSigner(mnemonic, config, cosmosConfig, ethNetworks)
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
		relayerRunnable = NewMessageRelayer(config, mintControllerMap, chainHealth.MessageRelayer)
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
