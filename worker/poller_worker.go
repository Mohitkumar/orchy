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
	worker                   *worker
	workerName               string
	client                   *client.RpcClient
	stop                     chan struct{}
	maxRetryBeforeResultPush int
	retryIntervalSecond      int
	wg                       *sync.WaitGroup
}

func (pw *pollerWorker) execute(task *api_v1.Action) *api_v1.ActionResult {
	result, err := pw.worker.Execute(util.ConvertFromProto(task.Data))
	var actionResult *api_v1.ActionResult
	if err != nil {
		actionResult = &api_v1.ActionResult{
			WorkflowName: task.WorkflowName,
			FlowId:       task.FlowId,
			ActionId:     task.ActionId,
			ActionName:   task.ActionName,
			Status:       api_v1.ActionResult_FAIL,
			RetryCount:   task.RetryCount,
		}
	} else {
		actionResult = &api_v1.ActionResult{
			WorkflowName: task.WorkflowName,
			FlowId:       task.FlowId,
			ActionId:     task.ActionId,
			Data:         util.ConvertToProto(result),
			Status:       api_v1.ActionResult_SUCCESS,
			RetryCount:   task.RetryCount,
		}
	}
	return actionResult

}

func (pw *pollerWorker) sendResponse(ctx context.Context, actionResult *api_v1.ActionResult) error {
	b := backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Duration(pw.retryIntervalSecond)*time.Second), uint64(pw.maxRetryBeforeResultPush))
	err := backoff.Retry(func() error {
		res, err := pw.client.GetApiClient().Push(ctx, actionResult)
		if err != nil {
			return err
		}
		logger.Debug("send result to server", zap.Bool("status", res.Status))
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
	req := &api_v1.ActionPollRequest{
		TaskType:  pw.worker.GetName(),
		BatchSize: int32(pw.worker.BatchSize()),
	}
	defer pw.wg.Done()
	for {
		select {
		case <-pw.stop:
			logger.Info("stopping worker", zap.String("worker", pw.workerName))
			ticker.Stop()
			return
		case <-ticker.C:
			actions, err := pw.client.GetApiClient().Poll(ctx, req)
			if err != nil {
				if e, ok := status.FromError(err); ok {
					switch e.Code() {
					case codes.Unavailable:
						logger.Error("server unavialable trying reconnect...")
					case codes.NotFound:
						logger.Debug("no action available", zap.Error(err))
					default:
						logger.Error("unexpected error", zap.Error(err))
					}
				}
			} else {
				for _, action := range actions.Actions {
					logger.Info("executing task", zap.String("flowId", action.FlowId), zap.String("task", action.ActionName))
					result := pw.execute(action)
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
	logger.Info("starting workers", zap.String("worker", pw.workerName))
	pw.wg.Add(1)
	ticker := time.NewTicker(pw.worker.GetPollInterval())
	go pw.workerLoop(ticker)
	logger.Info("started worker", zap.String("name", pw.workerName))
}

func (pw *pollerWorker) Stop() {
	pw.stop <- struct{}{}
	pw.client.Close()
}
