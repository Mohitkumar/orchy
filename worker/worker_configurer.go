package worker

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/worker/client"
)

type WorkerConfigurer struct {
	config       WorkerConfiguration
	actionPoller *actionPoller
	client       *client.RpcClient
	wg           sync.WaitGroup
}

func NewWorkerConfigurer(conf WorkerConfiguration) *WorkerConfigurer {
	client, err := client.NewRpcClient(conf.ServerUrl)
	if err != nil {
		panic(err)
	}
	actionPoller := newActionPoller(conf)

	wc := &WorkerConfigurer{
		config:       conf,
		actionPoller: actionPoller,
		client:       client,
	}
	return wc
}

func (wc *WorkerConfigurer) RegisterWorker(w *WorkerWrapper, name string, pollInterval time.Duration, batchSize int, numWorkers int) error {
	actionDef := &api.ActionDefinition{
		Name:              name,
		RetryCount:        int32(w.retryCount),
		RetryAfterSeconds: int32(w.retryAfterSeconds),
		RetryPolicy:       string(w.retryPolicy),
		TimeoutSeconds:    int32(w.timeoutSeconds),
	}
	ctx := context.Background()
	_, err := wc.client.GetApiClient().SaveActionDefinition(ctx, actionDef)
	if err != nil {
		return err
	}
	wc.actionPoller.registerWorker(newWorker(w.worker, name, pollInterval, batchSize), numWorkers)
	return nil
}

func (wc *WorkerConfigurer) Start() {
	wc.actionPoller.start(&wc.wg)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sigc
	wc.Stop()
}

func (wc *WorkerConfigurer) Stop() {
	wc.actionPoller.stop()
}
