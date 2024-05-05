package service

import (
	"fmt"

	"github.com/dan13ram/wpokt-oracle/models"
)

type EmptyRunner struct{}

func (e *EmptyRunner) Run() {
	fmt.Println("EmptyRunner run")
}

func (e *EmptyRunner) Status() models.RunnerStatus {
	return models.RunnerStatus{
		Enabled:     true,
		BlockHeight: 0,
	}
}
