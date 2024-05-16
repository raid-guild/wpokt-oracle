package service

type EmptyRunner struct{}

func (e *EmptyRunner) Run() {
}

func (e *EmptyRunner) Height() uint64 {
	return 0
}
