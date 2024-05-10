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

	monitorService RunnerServiceInterface
	signerService  RunnerServiceInterface
	relayerService RunnerServiceInterface

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
	if !x.monitorService.Enabled() && !x.signerService.Enabled() && !x.relayerService.Enabled() {
		log.Debugf("[%s] ChainService not enabled", x.Name())
		x.wg.Done()
		return
	}
	log.Infof("[%s] ChainService started", x.Name())

	var wg sync.WaitGroup

	if x.monitorService.Enabled() {
		wg.Add(1)
		go x.monitorService.Start(&wg)
	}

	if x.signerService.Enabled() {
		wg.Add(1)
		go x.signerService.Start(&wg)
	}

	if x.relayerService.Enabled() {
		wg.Add(1)
		go x.relayerService.Start(&wg)
	}

	<-x.stop

	log.Debugf("[%s] ChainService stopping", x.Name())

	if x.monitorService.Enabled() {
		x.monitorService.Stop()
	}
	if x.signerService.Enabled() {
		x.signerService.Stop()
	}
	if x.relayerService.Enabled() {
		x.relayerService.Stop()
	}

	wg.Wait()
	log.Infof("[%s] ChainService stopped", x.Name())

	x.wg.Done()
}

func (x *ChainService) Health() models.ServiceHealth {
	return models.ServiceHealth{
		Chain:          x.chain,
		MessageMonitor: x.monitorService.Status(),
		MessageSigner:  x.signerService.Status(),
		MessageRelayer: x.relayerService.Status(),
	}

}

func (x *ChainService) Stop() {
	log.Debugf("[%s] Stopping", x.Name())
	close(x.stop)
}

func NewChainService(
	chain models.Chain,
	monitorService RunnerServiceInterface,
	signerService RunnerServiceInterface,
	relayerService RunnerServiceInterface,
	wg *sync.WaitGroup,
) ChainServiceInterface {
	if chain.ChainName == "" || monitorService == nil || signerService == nil || relayerService == nil || wg == nil {
		log.Debug("[RUNNER] Invalid parameters")
		return nil
	}

	return &ChainService{
		chain:          chain,
		monitorService: monitorService,
		signerService:  signerService,
		relayerService: relayerService,
		wg:             wg,
		stop:           make(chan bool, 1),
	}
}

func NewEthereumChainService(
	config models.EthereumNetworkConfig,
	wg *sync.WaitGroup,
) ChainServiceInterface {
	monitorRunnerService := NewRunnerService(
		fmt.Sprintf("%s_Monitor", config.ChainName),
		&EmptyRunner{},
		config.MessageMonitor.Enabled,
		time.Duration(config.MessageMonitor.IntervalMS)*time.Millisecond,
	)
	signerRunnerService := NewRunnerService(
		fmt.Sprintf("%s_Signer", config.ChainName),
		&EmptyRunner{},
		config.MessageSigner.Enabled,
		time.Duration(config.MessageSigner.IntervalMS)*time.Millisecond,
	)
	relayerRunnerService := NewRunnerService(
		fmt.Sprintf("%s_Relayer", config.ChainName),
		&EmptyRunner{},
		config.MessageRelayer.Enabled,
		time.Duration(config.MessageRelayer.IntervalMS)*time.Millisecond,
	)

	return NewChainService(
		models.Chain{
			ChainName: config.ChainName,
			ChainId:   fmt.Sprintf("%d", config.ChainID),
			ChainType: models.ChainTypeEthereum,
		},
		monitorRunnerService,
		signerRunnerService,
		relayerRunnerService,
		wg,
	)
}
