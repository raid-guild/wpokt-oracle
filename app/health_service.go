package app

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/app/service"
	"github.com/dan13ram/wpokt-oracle/models"
)

type HealthService struct {
	interval time.Duration
	runner   *HealthCheckRunner
	stop     chan bool
	wg       *sync.WaitGroup
}

type HealthServiceInterface interface {
	Start(services []service.ChainServiceInterface)
	GetLastHealth() (models.Node, error)
	Stop()
}

func (x *HealthService) Start(
	services []service.ChainServiceInterface,
) {
	x.runner.AddServices(services)
	log.Infof("[HEALTH] HealthService started")
	stop := false
	for !stop {
		log.Infof("[HEALTH] Run started")

		x.runner.Run()

		log.Infof("[HEALTH] Run complete, next run in %s", x.interval)

		select {
		case <-x.stop:
			log.Infof("[HEALTH] HealthService stopped")
			x.wg.Done()
			stop = true
		case <-time.After(x.interval):
		}
	}
}

func (x *HealthService) GetLastHealth() (models.Node, error) {
	return x.runner.GetLastHealth()
}

func (x *HealthService) Stop() {
	log.Debugf("[HEALTH] Stopping")
	close(x.stop)
}

func NewHealthService(
	wg *sync.WaitGroup,
) HealthServiceInterface {
	interval := time.Duration(Config.HealthCheck.IntervalMS) * time.Millisecond
	runner := newHealthCheck()
	return &HealthService{
		runner:   runner,
		interval: interval,
		stop:     make(chan bool, 1),
		wg:       wg,
	}
}
