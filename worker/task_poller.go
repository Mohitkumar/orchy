package worker

import (
	"strconv"
	"sync"

	"github.com/mohitkumar/orchy/worker/client"
)

type nWorker struct {
	worker *worker
	num    int
}
type taskPoller struct {
	workers      []*nWorker
	pollerWorker []*pollerWorker
	config       WorkerConfiguration
}

func newTaskPoller(conf WorkerConfiguration) *taskPoller {
	return &taskPoller{
		config: conf,
	}
}

func (tp *taskPoller) registerWorker(worker *worker, numWorkers int) {
	tp.workers = append(tp.workers, &nWorker{worker: worker, num: numWorkers})
}

func (tp *taskPoller) start(wg *sync.WaitGroup) {
	for _, w := range tp.workers {
		for i := 0; i < w.num; i++ {
			client, err := client.NewRpcClient(tp.config.ServerUrl)
			if err != nil {
				panic(err)
			}
			pw := &pollerWorker{
				worker:                   w.worker,
				stop:                     make(chan struct{}),
				client:                   client,
				wg:                       wg,
				maxRetryBeforeResultPush: tp.config.MaxRetryBeforeResultPush,
				retryIntervalSecond:      tp.config.RetryIntervalSecond,
				workerName:               w.worker.GetName() + "_" + strconv.Itoa(i),
			}
			tp.pollerWorker = append(tp.pollerWorker, pw)
			pw.Start()
		}
	}
}

func (tp *taskPoller) stop() {
	for _, w := range tp.pollerWorker {
		w.Stop()
	}
}
