package health

import (
	"sync"
	"testing"
	"time"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
)

type mockHealthCheckRunnable struct {
	addServicesFunc   func(services []service.ChainService)
	runFunc           func()
	getLastHealthFunc func() (*models.Node, error)
}

func (m *mockHealthCheckRunnable) AddServices(services []service.ChainService) {
	if m.addServicesFunc != nil {
		m.addServicesFunc(services)
	}
}

func (m *mockHealthCheckRunnable) Run() {
	if m.runFunc != nil {
		m.runFunc()
	}
}

func (m *mockHealthCheckRunnable) GetLastHealth() (*models.Node, error) {
	if m.getLastHealthFunc != nil {
		return m.getLastHealthFunc()
	}
	return nil, nil
}

func TestHealthService_Start(t *testing.T) {
	var wg sync.WaitGroup
	mockRunnable := &mockHealthCheckRunnable{
		addServicesFunc: func(services []service.ChainService) {
			assert.Equal(t, 1, len(services))
		},
		runFunc: func() {},
	}

	healthService := &healthService{
		interval: time.Millisecond,
		runnable: mockRunnable,
		stop:     make(chan bool, 1),
		wg:       &wg,
		logger:   log.NewEntry(log.New()),
	}

	services := []service.ChainService{&mockChainService{}}

	wg.Add(1)
	go healthService.Start(services)
	time.Sleep(10 * time.Millisecond)
	healthService.Stop()
	wg.Wait()
}

func TestHealthService_GetLastHealth(t *testing.T) {
	mockRunnable := &mockHealthCheckRunnable{
		getLastHealthFunc: func() (*models.Node, error) {
			return &models.Node{}, nil
		},
	}

	healthService := &healthService{
		runnable: mockRunnable,
		logger:   log.NewEntry(log.New()),
	}

	node, err := healthService.GetLastHealth()
	assert.NoError(t, err)
	assert.NotNil(t, node)
}

func TestHealthService_Stop(t *testing.T) {
	var wg sync.WaitGroup
	mockRunnable := &mockHealthCheckRunnable{}

	healthService := &healthService{
		interval: time.Millisecond,
		runnable: mockRunnable,
		stop:     make(chan bool, 1),
		wg:       &wg,
		logger:   log.NewEntry(log.New()),
	}

	services := []service.ChainService{&mockChainService{}}

	wg.Add(1)
	go healthService.Start(services)
	time.Sleep(10 * time.Millisecond)
	healthService.Stop()
	wg.Wait()
}

func TestNewHealthService(t *testing.T) {
	var wg sync.WaitGroup
	config := models.Config{
		HealthCheck: models.HealthCheckConfig{
			IntervalMS: 1000,
		},
		Mnemonic: "infant apart enroll relief kangaroo patch awesome wagon trap feature armor approve",
		CosmosNetwork: models.CosmosNetworkConfig{
			MultisigPublicKeys: []string{
				"026892de2ec7fdf3125bc1bfd2ff2590d2c9ba756f98a05e9e843ac4d2a1acd4d9",
				"02faaaf0f385bb17381f36dcd86ab2486e8ff8d93440436496665ac007953076c2",
				"02cae233806460db75a941a269490ca5165a620b43241edb8bc72e169f4143a6df",
			},
		},
	}

	healthService := NewHealthService(config, &wg)
	assert.NotNil(t, healthService)
}
