package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	api_v1 "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/worker/client"
	"github.com/mohitkumar/orchy/worker/logger"
	"github.com/mohitkumar/orchy/worker/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type pollerWorker struct {
	worker                   Worker
	client                   *client.RpcClient
	stop                     chan struct{}
	maxRetryBeforeResultPush int
	retryIntervalSecond      int
	wg                       *sync.WaitGroup
	numWorker                int
}

func (pw *pollerWorker) execute(task *api_v1.Task) *api_v1.TaskResult {
	result, err := pw.worker.Execute(util.ConvertFromProto(task.Data))
	var taskResult *api_v1.TaskResult
	if err != nil {
		taskResult = &api_v1.TaskResult{
			WorkflowName: task.WorkflowName,
			FlowId:       task.FlowId,
			ActionId:     task.ActionId,
			TaskName:     task.TaskName,
			Status:       api_v1.TaskResult_FAIL,
			RetryCount:   task.RetryCount,
		}
	} else {
		taskResult = &api_v1.TaskResult{
			WorkflowName: task.WorkflowName,
			FlowId:       task.FlowId,
			ActionId:     task.ActionId,
			Data:         util.ConvertToProto(result),
			Status:       api_v1.TaskResult_SUCCESS,
			RetryCount:   task.RetryCount,
		}
	}
	return taskResult

}

func (pw *pollerWorker) sendResponse(ctx context.Context, taskResult *api_v1.TaskResult) error {
	b := backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Duration(pw.retryIntervalSecond)*time.Second), uint64(pw.maxRetryBeforeResultPush))
	err := backoff.Retry(func() error {
		res, err := pw.client.GetApiClient().Push(ctx, taskResult)
		if err != nil {
			return err
		}
		logger.Info("send result to server", zap.Bool("status", res.Status))
		if !res.Status {
			return fmt.Errorf("push task execution result failed")
		}
		return nil
	}, b)
	if err != nil {
		return err
	}
	return nil
}

func (pw *pollerWorker) workerLoop(ticker *time.Ticker) {
	ctx := context.WithValue(context.Background(), "worker", pw.worker.GetName())
	req := &api_v1.TaskPollRequest{
		TaskType:  pw.worker.GetName(),
		BatchSize: int32(pw.worker.BatchSize()),
	}
	defer pw.wg.Done()
	for {
		select {
		case <-pw.stop:
			ticker.Stop()
			return
		case <-ticker.C:
			tasks, err := pw.client.GetApiClient().Poll(ctx, req)
			if err != nil {
				if e, ok := status.FromError(err); ok {
					switch e.Code() {
					case codes.Unavailable:
						logger.Error("server unavialable trying reconnect...")
					case codes.NotFound:
						logger.Debug("not task available", zap.Error(err))
					default:
						logger.Error("unexpected error", zap.Error(err))
					}
				}
			} else {
				for _, task := range tasks.Tasks {
					logger.Info("executing task", zap.String("flowId", task.FlowId), zap.String("task", task.TaskName))
					result := pw.execute(task)
					err = pw.sendResponse(ctx, result)
					if err != nil {
						logger.Error("error sending task execution response to server", zap.String("taskType", pw.worker.GetName()))
					}
				}
			}
		}
	}
}
func (pw *pollerWorker) Start() {
	logger.Info("starting workers", zap.String("worker", pw.worker.GetName()), zap.Int("workerCount", pw.numWorker))
	for i := 0; i < pw.numWorker; i++ {
		pw.wg.Add(1)
		ticker := time.NewTicker(time.Duration(pw.worker.GetPollInterval()) * time.Second)
		go pw.workerLoop(ticker)
		logger.Info("started worker", zap.String("name", pw.worker.GetName()))
	}
}
