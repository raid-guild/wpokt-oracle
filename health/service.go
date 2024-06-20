package health

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dan13ram/wpokt-oracle/models"
	"github.com/dan13ram/wpokt-oracle/service"
)

type healthService struct {
	interval time.Duration
	runnable HealthCheckRunnable
	stop     chan bool
	wg       *sync.WaitGroup

	logger *log.Entry
}

type HealthService interface {
	Start(services []service.ChainService)
	GetLastHealth() (*models.Node, error)
	Stop()
}

func (x *healthService) Start(
	services []service.ChainService,
) {
	x.runnable.AddServices(services)
	x.logger.Infof("HealthService started")
	stop := false
	for !stop {
		x.logger.Infof("Run started")

		x.runnable.Run()

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

func (x *healthService) GetLastHealth() (*models.Node, error) {
	return x.runnable.GetLastHealth()
}

func (x *healthService) Stop() {
	x.logger.Debugf("HealthService stopping")
	close(x.stop)
}

func NewHealthService(
	config models.Config,
	wg *sync.WaitGroup,
) HealthService {
	interval := time.Duration(config.HealthCheck.IntervalMS) * time.Millisecond
	return &healthService{
		runnable: newHealthCheck(config),
		interval: interval,
		stop:     make(chan bool, 1),
		wg:       wg,

		logger: log.WithField("module", "health").WithField("service", "health"),
	}
}
