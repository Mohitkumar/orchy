package executor

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ Executor = new(systemActionExecutor)

type timeoutExecutor struct {
	diContainer *container.DIContiner
	shard       persistence.Shard
	wg          *sync.WaitGroup
	tw          *util.TickWorker
	stop        chan struct{}
}

func NewTimeoutExecutor(diContainer *container.DIContiner, shard persistence.Shard, wg *sync.WaitGroup) *timeoutExecutor {
	ex := &timeoutExecutor{
		diContainer: diContainer,
		shard:       shard,
		stop:        make(chan struct{}),
		wg:          wg,
	}
	ex.tw = util.NewTickWorker(ex.Name(), 1*time.Second, ex.stop, ex.handle, ex.wg)
	return ex
}

func (ex *timeoutExecutor) Name() string {
	return "timeout-executor-" + ex.shard.GetShardId()
}

func (ex *timeoutExecutor) Start() {
	if ex.IsRunning() {
		return
	}
	ex.tw.Start()
}

func (ex *timeoutExecutor) Stop() {
	if !ex.IsRunning() {
		return
	}
	ex.stop <- struct{}{}
}

func (ex *timeoutExecutor) IsRunning() bool {
	return ex.tw.IsRunning()
}

func (ex *timeoutExecutor) handle() {
	actions, err := ex.shard.PollTimeout()
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions.Actions {
		taskDef, err := ex.diContainer.GetMetadataStorage().GetActionDefinition(action.ActionName)
		if err != nil {
			logger.Error("task definition not found ", zap.String("taskName", action.ActionName), zap.Error(err))
			continue
		}
		if int(action.RetryCount) <= taskDef.RetryCount {
			var retryAfter time.Duration
			switch taskDef.RetryPolicy {
			case model.RETRY_POLICY_FIXED:
				retryAfter = time.Duration(taskDef.RetryAfterSeconds) * time.Second
			case model.RETRY_POLICY_BACKOFF:
				retryAfter = time.Duration(taskDef.RetryAfterSeconds*int(action.RetryCount)) * time.Second
			}
			action.RetryCount = action.RetryCount + 1
			ex.diContainer.GetClusterStorage().Retry(action, retryAfter)
		} else {
			logger.Error("task max retry exhausted, failing workflow", zap.Int("maxRetry", taskDef.RetryCount))
			flowMachine, err := flow.GetFlowStateMachine(action.WorkflowName, action.FlowId, ex.diContainer)
			if err != nil {
				logger.Error("task definition not found ", zap.String("taskName", action.ActionName), zap.Error(err))
				continue
			}
			flowMachine.MarkFailed()
		}
	}
}
