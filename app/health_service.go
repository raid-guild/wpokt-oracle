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

	logger *log.Entry
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
	x.logger.Infof("HealthService started")
	stop := false
	for !stop {
		x.logger.Infof("Run started")

		x.runner.Run()

		x.logger.Infof("Run complete, next run in %s", x.interval)

		select {
		case <-x.stop:
			x.logger.Infof("HealthService stopped")
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
	x.logger.Debugf("HealthService stopping")
	close(x.stop)
}

func NewHealthService(
	config models.Config,
	wg *sync.WaitGroup,
) HealthServiceInterface {
	interval := time.Duration(config.HealthCheck.IntervalMS) * time.Millisecond
	runner := newHealthCheck(config)
	return &HealthService{
		runner:   runner,
		interval: interval,
		stop:     make(chan bool, 1),
		wg:       wg,

		logger: log.WithField("module", "health").WithField("service", "health"),
	}
}
