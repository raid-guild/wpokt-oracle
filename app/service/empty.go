package service

import (
	"github.com/dan13ram/wpokt-oracle/models"
)

type EmptyRunner struct{}

func (e *EmptyRunner) Run() {
}

func (e *EmptyRunner) Status() models.RunnerStatus {
	return models.RunnerStatus{
		Enabled:     false,
		BlockHeight: 0,
	}
}
