package worker

import (
	"sync"

	"github.com/mohitkumar/orchy/worker/logger"
	"go.uber.org/zap"
)

type workerWithStopChannel struct {
	worker Worker
	stop   chan struct{}
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

func (tp *taskPoller) registerWorker(worker Worker) {
	stopc := make(chan struct{})
	tp.workers = append(tp.workers, &workerWithStopChannel{worker: worker, stop: stopc})
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
		}
		err = pw.Start()
		if err != nil {
			logger.Error("error starting worker ", zap.String("name", w.worker.GetName()), zap.Error(err))
		}
	}
}

func (tp *taskPoller) stop() {
	for _, w := range tp.workers {
		w.stop <- struct{}{}
	}
}
