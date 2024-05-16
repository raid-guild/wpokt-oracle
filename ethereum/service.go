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
	wg *sync.WaitGroup,
	nodeHealth models.Node,
) service.ChainServiceInterface {

	var chainHealth models.ChainServiceHealth
	for _, health := range nodeHealth.Health {
		if health.Chain.ChainID == fmt.Sprintf("%d", config.ChainID) && health.Chain.ChainType == models.ChainTypeCosmos {
			chainHealth = health
			break
		}
	}

	var monitorRunner service.Runner
	monitorRunner = &service.EmptyRunner{}
	if config.MessageMonitor.Enabled {
		monitorRunner = NewMessageMonitor(config, chainHealth.MessageMonitor)
	}
	monitorRunnerService := service.NewRunnerService(
		"monitor",
		monitorRunner,
		config.MessageMonitor.Enabled,
		time.Duration(config.MessageMonitor.IntervalMS)*time.Millisecond,
	)

	signerRunnerService := service.NewRunnerService(
		"signer",
		&service.EmptyRunner{},
		config.MessageSigner.Enabled,
		time.Duration(config.MessageSigner.IntervalMS)*time.Millisecond,
	)

	relayerRunnerService := service.NewRunnerService(
		"relayer",
		&service.EmptyRunner{},
		config.MessageRelayer.Enabled,
		time.Duration(config.MessageRelayer.IntervalMS)*time.Millisecond,
	)

	return service.NewChainService(
		util.ParseChain(config),
		monitorRunnerService,
		signerRunnerService,
		relayerRunnerService,
		wg,
	)
}
