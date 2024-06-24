package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyRunner(t *testing.T) {
	runner := &EmptyRunnable{}
	runner.Run()

	assert.Equal(t, uint64(0), runner.Height())
}
