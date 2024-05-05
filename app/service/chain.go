package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
)

type ChainService struct {
	wg *sync.WaitGroup

	chain models.Chain

	monitorRunner RunnerServiceInterface
	signerRunner  RunnerServiceInterface
	relayerRunner RunnerServiceInterface

	stop chan bool
}

type ChainServiceInterface interface {
	Start()
	Stop()
	Health() models.ServiceHealth
}

func (x *ChainService) Name() string {
	return strings.ToUpper(x.chain.ChainName)
}

func (x *ChainService) Start() {
	log.Infof("[%s] ChainService started", x.Name())

	var wg sync.WaitGroup
	wg.Add(3)

	go x.monitorRunner.Start(&wg)
	go x.signerRunner.Start(&wg)
	go x.relayerRunner.Start(&wg)

	<-x.stop

	log.Debugf("[%s] ChainService stopping", x.Name())
	x.monitorRunner.Stop()
	x.signerRunner.Stop()
	x.relayerRunner.Stop()

	wg.Wait()
	log.Infof("[%s] ChainService stopped", x.Name())

	x.wg.Done()
}

func (x *ChainService) Health() models.ServiceHealth {
	return models.ServiceHealth{
		Chain:          x.chain,
		MessageMonitor: x.monitorRunner.Status(),
		MessageSigner:  x.signerRunner.Status(),
		MessageRelayer: x.relayerRunner.Status(),
	}

}

func (x *ChainService) Stop() {
	log.Debugf("[%s] Stopping", x.Name())
	close(x.stop)
}

func NewChainService(
	chain models.Chain,
	monitorRunner RunnerServiceInterface,
	signerRunner RunnerServiceInterface,
	relayerRunner RunnerServiceInterface,
	wg *sync.WaitGroup,
) ChainServiceInterface {
	if chain.ChainName == "" || monitorRunner == nil || signerRunner == nil || relayerRunner == nil || wg == nil {
		log.Debug("[RUNNER] Invalid parameters")
		return nil
	}

	return &ChainService{
		chain:         chain,
		monitorRunner: monitorRunner,
		signerRunner:  signerRunner,
		relayerRunner: relayerRunner,
		wg:            wg,
		stop:          make(chan bool, 1),
	}
}

func NewEthereumChainService(
	config models.EthereumNetworkConfig,
	wg *sync.WaitGroup,
) ChainServiceInterface {
	monitorRunner := NewRunnerService(
		fmt.Sprintf("%s_Monitor", config.ChainName),
		&EmptyRunner{},
		time.Duration(config.MessageMonitor.IntervalMS)*time.Millisecond,
	)
	signerRunner := NewRunnerService(
		fmt.Sprintf("%s_Signer", config.ChainName),
		&EmptyRunner{},
		time.Duration(config.MessageSigner.IntervalMS)*time.Millisecond,
	)
	relayerRunner := NewRunnerService(
		fmt.Sprintf("%s_Relayer", config.ChainName),
		&EmptyRunner{},
		time.Duration(config.MessageRelayer.IntervalMS)*time.Millisecond,
	)

	return NewChainService(
		models.Chain{
			ChainName: config.ChainName,
			ChainId:   fmt.Sprintf("%d", config.ChainId),
			ChainType: models.ChainTypeEthereum,
		},
		monitorRunner,
		signerRunner,
		relayerRunner,
		wg,
	)
}

func NewCosmosChainService(
	config models.CosmosNetworkConfig,
	wg *sync.WaitGroup,
) ChainServiceInterface {
	monitorRunner := NewRunnerService(
		fmt.Sprintf("%s_Monitor", config.ChainName),
		&EmptyRunner{},
		time.Duration(config.MessageMonitor.IntervalMS)*time.Millisecond,
	)
	signerRunner := NewRunnerService(
		fmt.Sprintf("%s_Signer", config.ChainName),
		&EmptyRunner{},
		time.Duration(config.MessageSigner.IntervalMS)*time.Millisecond,
	)
	relayerRunner := NewRunnerService(
		fmt.Sprintf("%s_Relayer", config.ChainName),
		&EmptyRunner{},
		time.Duration(config.MessageRelayer.IntervalMS)*time.Millisecond,
	)

	return NewChainService(
		models.Chain{
			ChainName: config.ChainName,
			ChainId:   config.ChainId,
			ChainType: models.ChainTypeCosmos,
		},
		monitorRunner,
		signerRunner,
		relayerRunner,
		wg,
	)
}
