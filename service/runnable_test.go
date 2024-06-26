package service

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockRunnable struct {
	height uint64
}

func (m *mockRunnable) Run() {
	m.height++
}

func (m *mockRunnable) Height() uint64 {
	return m.height
}

func TestRunnerService_Enabled(t *testing.T) {
	r := &runnerService{enabled: true}
	assert.True(t, r.Enabled())

	r = &runnerService{enabled: false}
	assert.False(t, r.Enabled())
}

func TestRunnerService_Start_Disabled(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	r := &runnerService{
		enabled: false,
		logger:  log.NewEntry(log.New()),
	}
	r.Start(&wg)

	assert.NotNil(t, r)
	wg.Wait()
}

func TestRunnerService_Start_NilRunnable(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	r := &runnerService{
		enabled:  true,
		runnable: nil,
		logger:   log.NewEntry(log.New()),
	}
	r.Start(&wg)

	assert.NotNil(t, r)
	wg.Wait()
}

func TestRunnerService_Start_And_Stop(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	runnable := &mockRunnable{}
	r := &runnerService{
		enabled:  true,
		runnable: runnable,
		interval: 1 * time.Second,
		stop:     make(chan bool, 1),
		logger:   log.NewEntry(log.New()),
	}
	go r.Start(&wg)

	time.Sleep(2 * time.Second)
	r.Stop()

	wg.Wait()
	assert.GreaterOrEqual(t, runnable.height, uint64(1))
}

func TestRunnerService_Status(t *testing.T) {
	runnable := &mockRunnable{}
	r := &runnerService{
		enabled:  true,
		runnable: runnable,
		interval: 1 * time.Second,
		status: models.RunnerServiceStatus{
			Name:        "TestService",
			LastRunAt:   time.Now(),
			NextRunAt:   time.Now().Add(1 * time.Second),
			Enabled:     true,
			BlockHeight: 1,
		},
		logger: log.NewEntry(log.New()),
	}

	status := r.Status()
	assert.NotNil(t, status)
	assert.Equal(t, "TestService", status.Name)

	r.enabled = false
	status = r.Status()
	assert.Nil(t, status)
}

func TestRunnerService_UpdateStatus(t *testing.T) {
	r := &runnerService{
		name:     "TestService",
		enabled:  true,
		interval: 1 * time.Second,
		logger:   log.NewEntry(log.New()),
	}
	r.updateStatus(10)

	status := r.Status()
	assert.NotNil(t, status)
	assert.Equal(t, uint64(10), status.BlockHeight)
	assert.Equal(t, "TestService", status.Name)
	assert.True(t, status.Enabled)
}

func TestNewRunnerService(t *testing.T) {
	runnable := &mockRunnable{}
	chain := models.Chain{
		ChainName: "TestChain",
		ChainID:   "1",
	}

	r := NewRunnerService("TestService", runnable, true, 1*time.Second, chain)
	assert.NotNil(t, r)
	assert.Equal(t, "TESTSERVICE", r.(*runnerService).name)
	assert.True(t, r.Enabled())

	defer func() { log.StandardLogger().ExitFunc = nil }()
	log.StandardLogger().ExitFunc = func(num int) { panic(fmt.Sprintf("exit %d", num)) }

	assert.Panics(t, func() {
		NewRunnerService("TestService", nil, true, 1*time.Second, chain)
	})

	assert.Panics(t, func() {
		NewRunnerService("TestService", runnable, true, 0, chain)
	})
}
