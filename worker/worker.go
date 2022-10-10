package worker

import "time"

type worker struct {
	fn           func(map[string]any) (map[string]any, error)
	name         string
	pollInterval time.Duration
	batchSize    int
}

func newWorker(fn func(map[string]any) (map[string]any, error), name string, pollInterval time.Duration, batchSize int) *worker {
	return &worker{
		fn:           fn,
		name:         name,
		pollInterval: pollInterval,
		batchSize:    batchSize,
	}
}

func (w *worker) Execute(input map[string]any) (map[string]any, error) {
	return w.fn(input)
}
func (w *worker) GetName() string {
	return w.name
}
func (w *worker) GetPollInterval() time.Duration {
	return w.pollInterval
}
func (w *worker) BatchSize() int {
	return w.batchSize
}
