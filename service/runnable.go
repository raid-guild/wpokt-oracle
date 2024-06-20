package service

import (
	"strings"
	"sync"
	"time"

	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
)

type Runnable interface {
	Run()
	Height() uint64
}

type RunnerService interface {
	Start(wg *sync.WaitGroup)
	Enabled() bool
	Status() *models.RunnerServiceStatus
	Stop()
}

type runnerService struct {
	name string

	enabled  bool
	runnable Runnable
	interval time.Duration

	stop chan bool

	statusMu sync.RWMutex
	status   models.RunnerServiceStatus

	logger *log.Entry
}

func (x *runnerService) Enabled() bool {
	return x.enabled
}

func (x *runnerService) Start(wg *sync.WaitGroup) {
	if !x.enabled {
		x.logger.Debugf("RunnerService is disabled")
		wg.Done()
		return
	}
	if x.runnable == nil {
		x.logger.Debugf("RunnerService not started, runner is nil")
		wg.Done()
		return
	}

	x.logger.Infof("RunnerService started")
	stop := false
	for !stop {
		x.logger.Infof("Run started")

		x.runnable.Run()

		x.updateStatus(x.runnable.Height())

		x.logger.Infof("Run complete, next run in %s", x.interval)

		select {
		case <-x.stop:
			x.logger.Infof("RunnerService stopped")
			wg.Done()
			stop = true
		case <-time.After(x.interval):
		}
	}
}

func (x *runnerService) Status() *models.RunnerServiceStatus {
	x.statusMu.RLock()
	defer x.statusMu.RUnlock()

	if !x.enabled {
		return nil
	}

	statusCopy := x.status

	return &statusCopy
}

func (x *runnerService) updateStatus(blockHeight uint64) {
	x.statusMu.Lock()
	defer x.statusMu.Unlock()

	lastRunAt := time.Now()

	x.status = models.RunnerServiceStatus{
		Name:        x.name,
		LastRunAt:   lastRunAt,
		NextRunAt:   lastRunAt.Add(x.interval),
		Enabled:     x.enabled,
		BlockHeight: blockHeight,
	}
}

func (x *runnerService) Stop() {
	x.logger.Debugf("RunnerService stopping")
	close(x.stop)
}

func NewRunnerService(
	name string,
	runnable Runnable,
	enabled bool,
	interval time.Duration,
	chain models.Chain,
) RunnerService {
	logger := log.
		WithField("module", "service").
		WithField("service", "runner").
		WithField("name", strings.ToLower(name)).
		WithField("chain_name", strings.ToLower(chain.ChainName)).
		WithField("chain_id", strings.ToLower(chain.ChainID))

	if (runnable == nil) || (interval == 0) {
		logger.
			Debug("Invalid parameters")
		return nil
	}

	return &runnerService{
		name:     strings.ToUpper(name),
		runnable: runnable,
		enabled:  enabled,
		interval: interval,
		stop:     make(chan bool, 1),
		status:   models.RunnerServiceStatus{},
		logger:   logger,
	}
}
