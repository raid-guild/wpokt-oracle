package app

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type HealthService struct {
	interval time.Duration

	runner HealthCheckRunner
	name   string

	stop chan bool
}

func (x *HealthService) Start(wg *sync.WaitGroup) {
	log.Infof("[%s] HealthService started", x.name)
	stop := false
	for !stop {
		log.Infof("[%s] Run started", x.name)

		x.runner.Run()

		log.Infof("[%s] Run complete, next run in %s", x.name, x.interval)

		select {
		case <-x.stop:
			log.Infof("[%s] HealthService stopped", x.name)
			wg.Done()
			stop = true
		case <-time.After(x.interval):
		}
	}
}

func (x *HealthService) Stop() {
	log.Debugf("[%s] Stopping", x.name)
	close(x.stop)
}

func NewHealthService(
	runner Runner,
	interval time.Duration,
) HealthServiceInterface {
	if (runner == nil) || (interval == 0) {
		log.Debug("[RUNNER] Invalid parameters")
		return nil
	}

	return &HealthService{
		me:       "HEALTH",
		runner:   runner,
		interval: interval,
		stop:     make(chan bool, 1),
	}
}
