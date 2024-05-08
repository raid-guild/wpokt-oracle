package cosmos

import (
	"fmt"
	"sync"
	"time"

	"github.com/dan13ram/wpokt-oracle/app/service"
	"github.com/dan13ram/wpokt-oracle/models"
)

func NewCosmosChainService(
	config models.CosmosNetworkConfig,
	wg *sync.WaitGroup,
) service.ChainServiceInterface {
	var monitorRunner service.Runner
	monitorRunner = &service.EmptyRunner{}
	if config.MessageMonitor.Enabled {
		monitorRunner = NewMessageMonitor(config, models.ServiceHealth{})
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
		models.Chain{
			ChainName: config.ChainName,
			ChainId:   config.ChainId,
			ChainType: models.ChainTypeCosmos,
		},
		monitorRunnerService,
		signerRunnerService,
		relayerRunnerService,
		wg,
	)
}
