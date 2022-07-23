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
}

func (pw *pollerWorker) execute(task *api_v1.Task) *api_v1.TaskResult {
	result, err := pw.worker.Execute(util.ConvertFromProto(task.Data))
	var taskResult *api_v1.TaskResult
	if err != nil {
		taskResult = &api_v1.TaskResult{
			WorkflowName: task.WorkflowName,
			FlowId:       task.FlowId,
			ActionId:     task.ActionId,
			Status:       api_v1.TaskResult_FAIL,
		}
	} else {
		taskResult = &api_v1.TaskResult{
			WorkflowName: task.WorkflowName,
			FlowId:       task.FlowId,
			ActionId:     task.ActionId,
			Data:         util.ConvertToProto(result),
			Status:       api_v1.TaskResult_SUCCESS,
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
func (pw *pollerWorker) Start() error {
	ctx := context.Background()
	req := &api_v1.TaskPollRequest{
		TaskType: pw.worker.GetName(),
	}
	taskStream, err := pw.client.GetApiClient().PollStream(ctx, req)
	if err != nil {
		return err
	}
	pw.wg.Add(1)
	go func() {
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
							logger.Error("server not running reconnecting...")
							pw.client.Refresh()
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
	}()
	return nil
}
