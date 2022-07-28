package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	api_v1 "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/worker/logger"
	"github.com/mohitkumar/orchy/worker/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type pollerWorker struct {
	worker                   Worker
	client                   *client
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
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.Unavailable:
				pw.client.Refresh()
			}
		}
		if err != nil {
			return err
		}
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

func (pw *pollerWorker) workerLoop(errorChan chan<- error, ctx context.Context, taskStream api_v1.TaskService_PollStreamClient) {
	defer pw.wg.Done()
	for {
		select {
		case <-pw.stop:
			return
		default:
			task, err := taskStream.Recv()
			if err != nil {
				if e, ok := status.FromError(err); ok {
					switch e.Code() {
					case codes.Unavailable:
						errorChan <- err
						return
					default:
						logger.Error("error", zap.Error(err))
					}
				}
			} else {
				result := pw.execute(task)
				err = pw.sendResponse(ctx, result)
				if err != nil {
					logger.Error("error sending task execution response to server", zap.String("taskType", pw.worker.GetName()))
				}
			}
		}
	}
}
func (pw *pollerWorker) Start() {
	logger.Info("starting workers", zap.String("worker", pw.worker.GetName()), zap.Int("workers", pw.numWorker))
	ctx := context.Background()
	req := &api_v1.TaskPollRequest{
		TaskType: pw.worker.GetName(),
	}
	taskStream, err := pw.client.GetApiClient().PollStream(ctx, req)
	if err != nil {
		logger.Error("error opening stream to server")
		return
	}
	errorCh := make(chan error, pw.numWorker)
	for i := 0; i < pw.numWorker; i++ {
		pw.wg.Add(1)
		go pw.workerLoop(errorCh, ctx, taskStream)
		logger.Info("started worker", zap.String("name", pw.worker.GetName()))
	}
	for {
		select {
		case <-pw.stop:
			return
		case err := <-errorCh:
			for err != nil {
				logger.Error("server is unavialble reconnecting worker", zap.String("name", pw.worker.GetName()), zap.Error(err))
				time.Sleep(1 * time.Second)
				err = pw.client.Refresh()
				taskStream, err := pw.client.GetApiClient().PollStream(ctx, req)
				if err == nil {
					logger.Info("server reconnect successfull", zap.String("worker", pw.worker.GetName()))
					pw.wg.Add(1)
					errorCh = make(chan error, pw.numWorker)
					go pw.workerLoop(errorCh, ctx, taskStream)
				}
			}
		}
	}
}
