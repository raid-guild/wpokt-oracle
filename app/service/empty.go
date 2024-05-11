package service

import (
	"github.com/dan13ram/wpokt-oracle/models"
)

type EmptyRunner struct{}

func (e *EmptyRunner) Run() {
}

func (e *EmptyRunner) Status() models.RunnerServiceStatus {
	return models.RunnerServiceStatus{
		Enabled:     false,
		BlockHeight: 0,
	}
}
