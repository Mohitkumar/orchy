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
type TaskPoller struct {
	workers []*workerWithStopChannel
	Config  WorkerConfiguration
	wg      *sync.WaitGroup
}

func NewTaskPoller(conf WorkerConfiguration, wg *sync.WaitGroup) *TaskPoller {
	return &TaskPoller{
		Config: conf,
		wg:     wg,
	}
}

func (tp *TaskPoller) RegisterWorker(worker Worker) {
	stopc := make(chan struct{})
	tp.workers = append(tp.workers, &workerWithStopChannel{worker: worker, stop: stopc})
}

func (tp *TaskPoller) Start() {
	for _, w := range tp.workers {
		client, err := newClient(tp.Config.ServerUrl)
		if err != nil {
			panic(err)
		}
		pw := &pollerWorker{
			worker:                   w.worker,
			stop:                     w.stop,
			client:                   client,
			wg:                       tp.wg,
			maxRetryBeforeResultPush: tp.Config.MaxRetryBeforeResultPush,
			retryIntervalSecond:      tp.Config.RetryIntervalSecond,
		}
		err = pw.Start()
		if err != nil {
			logger.Error("error starting worker ", zap.String("name", w.worker.GetName()), zap.Error(err))
		}
	}
}

func (tp *TaskPoller) Stop() {
	for _, w := range tp.workers {
		w.stop <- struct{}{}
	}
}
