package worker

import (
	"context"
	"sync"
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/worker/client"
)

type WorkerConfigurer struct {
	config       WorkerConfiguration
	actionPoller *actionPoller
	client       *client.RpcClient
	clusterConf  *client.ClusterConf
	wg           *sync.WaitGroup
}

func NewWorkerConfigurer(conf WorkerConfiguration, wg *sync.WaitGroup) *WorkerConfigurer {
	clusterConf := client.NewclusterConf(conf.ServerUrl, wg)
	actionPoller := newActionPoller(conf, clusterConf)
	client, err := client.NewRpcClient(conf.ServerUrl, clusterConf)
	if err != nil {
		panic(err)
	}
	wc := &WorkerConfigurer{
		config:       conf,
		actionPoller: actionPoller,
		client:       client,
		clusterConf:  clusterConf,
		wg:           wg,
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
	wc.clusterConf.Start()
	wc.actionPoller.start(wc.wg)
}

func (wc *WorkerConfigurer) Stop() {
	wc.actionPoller.stop()
	wc.clusterConf.Stop()
}
