package worker

import (
	"sync"
)

type workerWithStopChannel struct {
	worker     Worker
	stop       chan struct{}
	numWorkers int
}
type taskPoller struct {
	workers []*workerWithStopChannel
	config  WorkerConfiguration
	wg      *sync.WaitGroup
}

func NewTaskPoller(conf WorkerConfiguration, wg *sync.WaitGroup) *taskPoller {
	return &taskPoller{
		config: conf,
		wg:     wg,
	}
}

func (tp *taskPoller) registerWorker(worker Worker, numWorkers int) {
	stopc := make(chan struct{})
	tp.workers = append(tp.workers, &workerWithStopChannel{worker: worker, stop: stopc, numWorkers: numWorkers})
}

func (tp *taskPoller) start() {
	for _, w := range tp.workers {
		client, err := newClient(tp.config.ServerUrl)
		if err != nil {
			panic(err)
		}
		pw := &pollerWorker{
			worker:                   w.worker,
			stop:                     w.stop,
			client:                   client,
			wg:                       tp.wg,
			maxRetryBeforeResultPush: tp.config.MaxRetryBeforeResultPush,
			retryIntervalSecond:      tp.config.RetryIntervalSecond,
			numWorker:                w.numWorkers,
		}
		tp.wg.Add(1)
		go pw.Start()
	}
}

func (tp *taskPoller) stop() {
	for _, w := range tp.workers {
		w.stop <- struct{}{}
	}
}
