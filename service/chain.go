package service

import (
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/models"
)

type ChainService struct {
	wg *sync.WaitGroup

	chain models.Chain

	monitorService RunnerServiceInterface
	signerService  RunnerServiceInterface
	relayerService RunnerServiceInterface

	stop   chan bool
	logger *log.Entry
}

type ChainServiceInterface interface {
	Start()
	Stop()
	Health() models.ChainServiceHealth
}

func (x *ChainService) Name() string {
	return strings.ToUpper(x.chain.ChainName)
}

func (x *ChainService) Start() {
	if !x.monitorService.Enabled() && !x.signerService.Enabled() && !x.relayerService.Enabled() {
		x.logger.Debugf("ChainService not enabled")
		x.wg.Done()
		return
	}
	x.logger.Infof("ChainService started")

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

	x.logger.Debugf("ChainService stopping")

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
	x.logger.Infof("ChainService stopped")

	x.wg.Done()
}

func (x *ChainService) Health() models.ChainServiceHealth {

	return models.ChainServiceHealth{
		Chain:          x.chain,
		MessageMonitor: x.monitorService.Status(),
		MessageSigner:  x.signerService.Status(),
		MessageRelayer: x.relayerService.Status(),
	}

}

func (x *ChainService) Stop() {
	x.logger.Debugf("ChainService Stopping")
	close(x.stop)
}

func NewChainService(
	chain models.Chain,
	monitorService RunnerServiceInterface,
	signerService RunnerServiceInterface,
	relayerService RunnerServiceInterface,
	wg *sync.WaitGroup,
) ChainServiceInterface {
	logger := log.
		WithField("module", "service").
		WithField("service", "chain").
		WithField("chain_name", strings.ToLower(chain.ChainName)).
		WithField("chain_id", strings.ToLower(chain.ChainID))
	if chain.ChainName == "" || monitorService == nil || signerService == nil || relayerService == nil || wg == nil {
		logger.Debug("Invalid parameters")
		return nil
	}

	return &ChainService{
		chain:          chain,
		monitorService: monitorService,
		signerService:  signerService,
		relayerService: relayerService,
		wg:             wg,
		stop:           make(chan bool, 1),
		logger:     logger,
	}
}
