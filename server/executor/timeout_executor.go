package executor

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

type TimeoutExecutor struct {
	container *container.DIContiner
	wg        *sync.WaitGroup
	stop      chan struct{}
}

func NewTimeoutExecutor(container *container.DIContiner, wg *sync.WaitGroup) *TimeoutExecutor {
	return &TimeoutExecutor{
		container: container,
		stop:      make(chan struct{}),
		wg:        wg,
	}
}

func (ex *TimeoutExecutor) Name() string {
	return "timeout-executor"
}

func (ex *TimeoutExecutor) handle(msg *model.ActionExecutionRequest) error {
	flowMachine, err := flow.GetFlowStateMachine(msg.WorkflowName, msg.FlowId, ex.container)
	if err != nil {
		return err
	}
	err = flowMachine.ValidateExecutionRequest(msg.ActionId)
	if err != nil {
		logger.Debug("discarding timeout action execution request, action has already executed")
		return err
	}
	taskDef, err := ex.container.GetTaskDao().GetTask(msg.TaskName)
	if err != nil {
		logger.Error("task definition not found ", zap.String("taskName", msg.TaskName), zap.Error(err))
		return err
	}
	if msg.TryNumber <= taskDef.RetryCount {
		var retryAfter time.Duration
		switch taskDef.RetryPolicy {
		case model.RETRY_POLICY_FIXED:
			retryAfter = time.Duration(taskDef.RetryAfterSeconds) * time.Second
		case model.RETRY_POLICY_BACKOFF:
			retryAfter = time.Duration(taskDef.RetryAfterSeconds*int(msg.TryNumber)) * time.Second
		}
		msg.TryNumber = msg.TryNumber + 1

		data, _ := ex.container.ActionExecutionRequestEncDec.Encode(*msg)

		ex.container.GetTaskRetryQueue().PushWithDelay("retry-queue", msg.FlowId, retryAfter, data)
	} else {
		logger.Error("task max retry exhausted, failing workflow", zap.Int("maxRetry", taskDef.RetryCount))
		flowMachine.MarkFailed()
	}
	return nil
}
func (ex *TimeoutExecutor) Start() error {
	fn := func() {
		res, err := ex.container.GetTaskTimeoutQueue().Pop("timeout-queue")
		if err != nil {
			logger.Error("error while polling timeout queue", zap.Error(err))
			return
		}
		for _, r := range res {
			msg, err := ex.container.ActionExecutionRequestEncDec.Decode([]byte(r))
			if err != nil {
				logger.Error("can not decode action execution request")
				continue
			}
			err = ex.handle(msg)
			if err != nil {
				continue
			}
		}
	}
	tw := util.NewTickWorker("timeout-worker", 1, ex.stop, fn, ex.wg)
	tw.Start()
	logger.Info("timeout executor started")
	return nil
}

func (ex *TimeoutExecutor) Stop() error {
	ex.stop <- struct{}{}
	return nil
}
