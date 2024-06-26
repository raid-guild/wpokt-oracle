package service

import (
	"fmt"
	"sync"
	"testing"

	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockRunnerService struct {
	enabled bool
	status  models.RunnerServiceStatus
	stop    chan bool
}

func (m *mockRunnerService) Start(wg *sync.WaitGroup) {
	wg.Done()
}

func (m *mockRunnerService) Enabled() bool {
	return m.enabled
}

func (m *mockRunnerService) Status() *models.RunnerServiceStatus {
	if !m.enabled {
		return nil
	}
	return &m.status
}

func (m *mockRunnerService) Stop() {
	close(m.stop)
}

func TestChainService_Name(t *testing.T) {
	chain := models.Chain{ChainName: "TestChain"}
	cs := &chainService{chain: chain}
	assert.Equal(t, "TESTCHAIN", cs.Name())
}

func TestChainService_Start_Disabled(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	cs := &chainService{
		wg: &wg,
		chain: models.Chain{
			ChainName: "TestChain",
		},
		monitorService: &mockRunnerService{enabled: false},
		signerService:  &mockRunnerService{enabled: false},
		relayerService: &mockRunnerService{enabled: false},
		logger:         log.NewEntry(log.New()),
	}
	go cs.Start()

	wg.Wait()
}

func TestChainService_Start_And_Stop(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	cs := &chainService{
		wg: &wg,
		chain: models.Chain{
			ChainName: "TestChain",
		},
		monitorService: &mockRunnerService{enabled: true, stop: make(chan bool, 1)},
		signerService:  &mockRunnerService{enabled: true, stop: make(chan bool, 1)},
		relayerService: &mockRunnerService{enabled: true, stop: make(chan bool, 1)},
		stop:           make(chan bool, 1),
		logger:         log.NewEntry(log.New()),
	}

	go cs.Start()
	cs.Stop()

	wg.Wait()
}

func TestChainService_Health(t *testing.T) {
	cs := &chainService{
		chain: models.Chain{
			ChainName: "TestChain",
			ChainID:   "1",
		},
		monitorService: &mockRunnerService{enabled: true, status: models.RunnerServiceStatus{Name: "MonitorService"}},
		signerService:  &mockRunnerService{enabled: true, status: models.RunnerServiceStatus{Name: "SignerService"}},
		relayerService: &mockRunnerService{enabled: true, status: models.RunnerServiceStatus{Name: "RelayerService"}},
		logger:         log.NewEntry(log.New()),
	}

	health := cs.Health()
	assert.Equal(t, "TestChain", health.Chain.ChainName)
	assert.Equal(t, "1", health.Chain.ChainID)
	assert.Equal(t, "MonitorService", health.MessageMonitor.Name)
	assert.Equal(t, "SignerService", health.MessageSigner.Name)
	assert.Equal(t, "RelayerService", health.MessageRelayer.Name)
}

func TestNewChainService(t *testing.T) {

	var wg sync.WaitGroup
	chain := models.Chain{ChainName: "TestChain", ChainID: "1"}

	cs := NewChainService(chain, &mockRunnerService{}, &mockRunnerService{}, &mockRunnerService{}, &wg)
	assert.NotNil(t, cs)
	assert.Equal(t, "TESTCHAIN", cs.(*chainService).Name())

	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	assert.Panics(t, func() { NewChainService(models.Chain{}, nil, nil, nil, nil) })
}
