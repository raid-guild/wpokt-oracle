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
	Status() models.RunnerStatus
}

type RunnerServiceInterface interface {
	Start(wg *sync.WaitGroup)
	Enabled() bool
	Status() models.RunnerStatus
	Stop()
}

type RunnerService struct {
	name string

	enabled  bool
	runner   Runner
	interval time.Duration

	stop chan bool

	statusMu sync.RWMutex
	status   models.RunnerStatus
}

func (x *RunnerService) Enabled() bool {
	return x.enabled
}

func (x *RunnerService) Start(wg *sync.WaitGroup) {
	if x.enabled == false {
		log.Debugf("[%s] RunnerService is disabled", x.name)
		return
	}
	if x.runner == nil {
		log.Debugf("[%s] RunnerService not started, runner is nil", x.name)
		return
	}

	log.Infof("[%s] RunnerService started", x.name)
	stop := false
	for !stop {
		log.Infof("[%s] Run started", x.name)

		x.runner.Run()

		x.updateStatus(x.runner.Status())

		log.Infof("[%s] Run complete, next run in %s", x.name, x.interval)

		select {
		case <-x.stop:
			log.Infof("[%s] RunnerService stopped", x.name)
			wg.Done()
			stop = true
		case <-time.After(x.interval):
		}
	}
}

func (x *RunnerService) Status() models.RunnerStatus {
	x.statusMu.RLock()
	defer x.statusMu.RUnlock()

	return x.status
}

func (x *RunnerService) updateStatus(status models.RunnerStatus) {
	x.statusMu.Lock()
	defer x.statusMu.Unlock()

	lastRunAt := time.Now()

	x.status = models.RunnerStatus{
		Name:        x.name,
		LastRunAt:   lastRunAt,
		NextRunAt:   lastRunAt.Add(x.interval),
		Enabled:     status.Enabled,
		BlockHeight: status.BlockHeight,
	}
}

func (x *RunnerService) Stop() {
	log.Debugf("[%s] Stopping", x.name)
	close(x.stop)
}

func NewRunnerService(
	name string,
	runner Runner,
	enabled bool,
	interval time.Duration,
) RunnerServiceInterface {
	if (runner == nil) || (interval == 0) {
		log.Debug("[RUNNER] Invalid parameters")
		return nil
	}

	return &RunnerService{
		name:     strings.ToUpper(name),
		runner:   runner,
		enabled:  enabled,
		interval: interval,
		stop:     make(chan bool, 1),
		status:   models.RunnerStatus{},
	}
}
