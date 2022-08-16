package executor

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

type TimeoutExecutor struct {
	container     *container.DIContiner
	wg            *sync.WaitGroup
	stop          chan struct{}
	actionExector *ActionExecutor
}

func NewTimeoutExecutor(container *container.DIContiner, actionExector *ActionExecutor, wg *sync.WaitGroup) *TimeoutExecutor {
	return &TimeoutExecutor{
		container:     container,
		actionExector: actionExector,
		stop:          make(chan struct{}),
		wg:            wg,
	}
}

func (ex *TimeoutExecutor) Name() string {
	return "timeout-executor"
}

func (ex *TimeoutExecutor) Start() error {
	fn := func() {
		res, err := ex.container.GetTaskTimeoutQueue().Pop("timeout-queue")
		if err != nil {
			_, ok := err.(persistence.EmptyQueueError)
			if !ok {
				logger.Error("error while polling timeout queue", zap.Error(err))
			}
			return
		}
		for _, r := range res {
			msg, err := ex.container.ActionExecutionRequestEncDec.Decode([]byte(r))
			if err != nil {
				logger.Error("can not decode action execution request")
				continue
			}
			taskDef, err := ex.container.GetTaskDao().GetTask(msg.TaskName)
			if err != nil {
				logger.Error("task definition not found ", zap.String("taskName", msg.TaskName), zap.Error(err))
				continue
			}
			if msg.TryNumber <= taskDef.RetryCount {
				var retryAfter time.Duration
				switch taskDef.RetryPolicy {
				case model.RETRY_POLICY_FIXED:
					retryAfter = time.Duration(taskDef.RetryAfterSeconds) * time.Second
				case model.RETRY_POLICY_BACKOFF:
					retryAfter = time.Duration(taskDef.RetryAfterSeconds*int(msg.TryNumber)) * time.Second
				}
				req := model.ActionExecutionRequest{
					WorkflowName: msg.WorkflowName,
					ActionId:     msg.ActionId,
					FlowId:       msg.FlowId,
					TryNumber:    msg.TryNumber + 1,
				}
				data, _ := ex.container.ActionExecutionRequestEncDec.Encode(req)

				ex.container.GetTaskRetryQueue().PushWithDelay("retry-queue", retryAfter, data)
			} else {
				logger.Error("task max retry exhausted, failing workflow", zap.Int("maxRetry", taskDef.RetryCount))
				flowCtx, err := ex.container.GetFlowDao().GetFlowContext(msg.WorkflowName, msg.FlowId)
				if err != nil {
					logger.Error("could not find workflow", zap.Error(err))
					return
				}
				flowCtx.State = model.FAILED
				err = ex.container.GetFlowDao().SaveFlowContext(msg.WorkflowName, msg.FlowId, flowCtx)
				if err != nil {
					logger.Error("could not save workflow context", zap.Error(err))
				}
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
