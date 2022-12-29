package executor

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ Executor = new(systemActionExecutor)

type retryExecutor struct {
	diContainer *container.DIContiner
	shard       persistence.Shard
	tw          *util.TickWorker
	wg          *sync.WaitGroup
	stop        chan struct{}
}

func NewRetryExecutor(diContainer *container.DIContiner, shard persistence.Shard, wg *sync.WaitGroup) *retryExecutor {
	ex := &retryExecutor{
		diContainer: diContainer,
		shard:       shard,
		stop:        make(chan struct{}),
		wg:          wg,
	}
	ex.tw = util.NewTickWorker(ex.Name(), 1*time.Second, ex.stop, ex.handle, ex.wg)
	return ex
}

func (ex *retryExecutor) Name() string {
	return "retryexecutor-" + ex.shard.GetShardId()
}

func (ex *retryExecutor) Start() {
	if ex.IsRunning() {
		return
	}
	ex.tw.Start()
}

func (ex *retryExecutor) IsRunning() bool {
	return ex.tw.IsRunning()
}

func (ex *retryExecutor) Stop() {
	if !ex.IsRunning() {
		return
	}
	ex.stop <- struct{}{}
}

func (ex *retryExecutor) handle() {
	actions, err := ex.shard.PollRetry()
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions.Actions {
		flowMachine, err := flow.GetFlowStateMachine(action.WorkflowName, action.FlowId, ex.diContainer)
		if err != nil {
			logger.Debug("error in executing workflow", zap.String("wfName", action.WorkflowName), zap.String("flowId", action.FlowId), zap.Error(err))
			continue
		}
		if flowMachine.IsCompleted() {
			logger.Debug("flow already complted", zap.String("wfName", action.WorkflowName), zap.String("flowId", action.FlowId))
			continue
		}
		err = flowMachine.DispatchAction(int(action.ActionId), int(action.RetryCount))
		if err != nil {
			logger.Error("error in retrying workflow", zap.String("wfName", action.WorkflowName), zap.String("flowId", action.FlowId), zap.Error(err))
		}
	}
}
