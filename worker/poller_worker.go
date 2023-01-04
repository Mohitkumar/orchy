package worker

import (
	"context"
	"fmt"
	"io"
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

func (pw *pollerWorker) execute(action *api_v1.Action) *api_v1.ActionResult {
	result, err := pw.worker.Execute(util.ConvertFromProto(action.Data))
	var actionResult *api_v1.ActionResult
	if err != nil {
		actionResult = &api_v1.ActionResult{
			WorkflowName: action.WorkflowName,
			FlowId:       action.FlowId,
			ActionId:     action.ActionId,
			ActionName:   action.ActionName,
			Status:       api_v1.ActionResult_FAIL,
			RetryCount:   action.RetryCount,
		}
	} else {
		actionResult = &api_v1.ActionResult{
			WorkflowName: action.WorkflowName,
			FlowId:       action.FlowId,
			ActionId:     action.ActionId,
			Data:         util.ConvertToProto(result),
			Status:       api_v1.ActionResult_SUCCESS,
			RetryCount:   action.RetryCount,
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
			return fmt.Errorf("push action execution result failed")
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
		ActionType: pw.worker.GetName(),
		BatchSize:  int32(pw.worker.BatchSize()),
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
					logger.Debug("executing action", zap.String("flowId", action.FlowId), zap.String("action", action.ActionName))
					result := pw.execute(action)
					err = pw.sendResponse(ctx, result)
					if err != nil {
						logger.Error("error sending action execution response to server", zap.String("actionType", pw.worker.GetName()))
					}
				}
			}
		}
	}
}
func (pw *pollerWorker) Start() {
	logger.Info("starting workers", zap.String("worker", pw.workerName))
	//pw.wg.Add(1)
	//ticker := time.NewTicker(pw.worker.GetPollInterval())
	go pw.workerLoopSteram()
}

func (pw *pollerWorker) Stop() {
	pw.stop <- struct{}{}
	pw.client.Close()
}

func (pw *pollerWorker) workerLoopSteram() error {
	ctx := context.WithValue(context.Background(), "worker", pw.worker.GetName())
	req := &api_v1.ActionPollRequest{
		ActionType: pw.worker.GetName(),
		BatchSize:  int32(pw.worker.BatchSize()),
	}
	stream, err := pw.client.GetApiClient().PollStream(ctx, req)
	if err != nil {
		return err
	}
	completed := make(chan bool)
	go func() {
		for {
			fmt.Println("running")
			action, err := stream.Recv()
			if err == io.EOF {
				completed <- true
			}
			if err != nil {
				logger.Error("cannot receive", zap.Error(err))
				completed <- true
			}
			result := pw.execute(action)
			err = pw.sendResponse(ctx, result)
			if err != nil {
				logger.Error("error sending action execution response to server", zap.String("actionType", pw.worker.GetName()))
			}
		}
	}()
	<-completed
	return nil
}

func (pw *pollerWorker) consumeStream(ctx context.Context, stream api_v1.ActionService_PollStreamClient, completed chan<- bool) {
	for {
		action, err := stream.Recv()
		if err == io.EOF {
			completed <- true
			return
		}
		if err != nil {
			logger.Error("cannot receive", zap.Error(err))
			completed <- true
			return
		}
		result := pw.execute(action)
		err = pw.sendResponse(ctx, result)
		if err != nil {
			logger.Error("error sending action execution response to server", zap.String("actionType", pw.worker.GetName()))
		}
	}
}
