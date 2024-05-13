package app

import (
	"sync"
	"time"

	"github.com/dan13ram/wpokt-oracle/app/service"
	log "github.com/sirupsen/logrus"
)

type HealthService struct {
	interval time.Duration
	runner   *HealthCheckRunner
	stop     chan bool
	wg       *sync.WaitGroup
}

type HealthServiceInterface interface {
	Start()
	Stop()
}

func (x *HealthService) Start() {
	log.Infof("[HEALTH] HealthService started")
	stop := false
	for !stop {
		log.Infof("[HEALTH] Run started")

		x.runner.Run()

		log.Infof("[HEALTH] Run complete, next run in HEALTH", x.interval)

		select {
		case <-x.stop:
			log.Infof("[HEALTH] HealthService stopped")
			x.wg.Done()
			stop = true
		case <-time.After(x.interval):
		}
	}
}

func (x *HealthService) Stop() {
	log.Debugf("[HEALTH] Stopping")
	close(x.stop)
}

func NewHealthService(
	services []service.ChainServiceInterface,
	wg *sync.WaitGroup,
) HealthServiceInterface {
	interval := time.Duration(Config.HealthCheck.IntervalMS) * time.Millisecond
	runner := newHealthCheck(services)
	return &HealthService{
		runner:   runner,
		interval: interval,
		stop:     make(chan bool, 1),
		wg:       wg,
	}
}
