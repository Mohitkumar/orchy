package worker

import (
	"context"
	"sync"

	api "github.com/mohitkumar/orchy/api/v1"
)

type WorkerConfigurer struct {
	config     WorkerConfiguration
	taskPoller *taskPoller
	client     *client
}

func NewWorkerConfigurer(conf WorkerConfiguration, wg *sync.WaitGroup) *WorkerConfigurer {
	client, err := newClient(conf.ServerUrl)
	if err != nil {
		panic(err)
	}
	taskPoller := NewTaskPoller(conf, wg)

	wc := &WorkerConfigurer{
		config:     conf,
		taskPoller: taskPoller,
		client:     client,
	}
	return wc
}

func (wc *WorkerConfigurer) RegisterWorker(w WorkerWrapper, numWorkers int) error {
	taskDef := &api.TaskDef{
		Name:              w.GetName(),
		RetryCount:        int32(w.retryCount),
		RetryAfterSeconds: int32(w.retryAfterSeconds),
		RetryPolicy:       string(w.retryPolicy),
		TimeoutSeconds:    int32(w.timeoutSeconds),
	}
	ctx := context.Background()
	_, err := wc.client.GetApiClient().SaveTaskDef(ctx, taskDef)
	if err != nil {
		return err
	}
	wc.taskPoller.registerWorker(w.Worker, numWorkers)
	return nil
}

func (wc *WorkerConfigurer) Start() {
	wc.taskPoller.start()
}

func (wc *WorkerConfigurer) Stop() {
	wc.taskPoller.stop()
}
