package service

type EmptyRunnable struct{}

func (e *EmptyRunnable) Run() {
}

func (e *EmptyRunnable) Height() uint64 {
	return 0
}
