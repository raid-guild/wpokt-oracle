package service

import (
	"strings"
	"sync"
	"time"

	"github.com/dan13ram/wpokt-oracle/models"
	log "github.com/sirupsen/logrus"
)

type Runner interface {
	Run()
	Height() uint64
}

type RunnerServiceInterface interface {
	Start(wg *sync.WaitGroup)
	Enabled() bool
	Status() *models.RunnerServiceStatus
	Stop()
}

type RunnerService struct {
	name string

	enabled  bool
	runner   Runner
	interval time.Duration

	stop chan bool

	statusMu sync.RWMutex
	status   models.RunnerServiceStatus

	logger *log.Entry
}

func (x *RunnerService) Enabled() bool {
	return x.enabled
}

func (x *RunnerService) Start(wg *sync.WaitGroup) {
	if !x.enabled {
		x.logger.Debugf("RunnerService is disabled")
		wg.Done()
		return
	}
	if x.runner == nil {
		x.logger.Debugf("RunnerService not started, runner is nil")
		wg.Done()
		return
	}

	x.logger.Infof("RunnerService started")
	stop := false
	for !stop {
		x.logger.Infof("Run started")

		x.runner.Run()

		x.updateStatus(x.runner.Height())

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

func (x *RunnerService) Status() *models.RunnerServiceStatus {
	x.statusMu.RLock()
	defer x.statusMu.RUnlock()

	if !x.enabled {
		return nil
	}

	statusCopy := x.status

	return &statusCopy
}

func (x *RunnerService) updateStatus(blockHeight uint64) {
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

func (x *RunnerService) Stop() {
	x.logger.Debugf("RunnerService stopping")
	close(x.stop)
}

func NewRunnerService(
	name string,
	runner Runner,
	enabled bool,
	interval time.Duration,
) RunnerServiceInterface {
	logger := log.
		WithField("module", "service").
		WithField("service", "runner").
		WithField("name", strings.ToLower(name))
	if (runner == nil) || (interval == 0) {
		logger.
			Debug("Invalid parameters")
		return nil
	}

	return &RunnerService{
		name:     strings.ToUpper(name),
		runner:   runner,
		enabled:  enabled,
		interval: interval,
		stop:     make(chan bool, 1),
		status:   models.RunnerServiceStatus{},
		logger:   logger,
	}
}
