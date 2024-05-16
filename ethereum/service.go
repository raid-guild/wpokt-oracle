package ethereum

import (
	"fmt"
	"sync"
	"time"

	"github.com/dan13ram/wpokt-oracle/app/service"
	"github.com/dan13ram/wpokt-oracle/ethereum/util"
	"github.com/dan13ram/wpokt-oracle/models"
)

func NewEthereumChainService(
	config models.EthereumNetworkConfig,
	wg *sync.WaitGroup,
	nodeHealth models.Node,
) service.ChainServiceInterface {

	var chainHealth models.ChainServiceHealth
	for _, health := range nodeHealth.Health {
		if health.Chain.ChainID == string(config.ChainID) && health.Chain.ChainType == models.ChainTypeCosmos {
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
		fmt.Sprintf("%s_Monitor", config.ChainName),
		monitorRunner,
		config.MessageMonitor.Enabled,
		time.Duration(config.MessageMonitor.IntervalMS)*time.Millisecond,
	)

	signerRunnerService := service.NewRunnerService(
		fmt.Sprintf("%s_Signer", config.ChainName),
		&service.EmptyRunner{},
		config.MessageSigner.Enabled,
		time.Duration(config.MessageSigner.IntervalMS)*time.Millisecond,
	)

	relayerRunnerService := service.NewRunnerService(
		fmt.Sprintf("%s_Relayer", config.ChainName),
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
