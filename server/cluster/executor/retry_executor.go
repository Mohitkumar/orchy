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
	wg          *sync.WaitGroup
	stop        chan struct{}
}

func NewRetryExecutor(diContainer *container.DIContiner, shard persistence.Shard, wg *sync.WaitGroup) *retryExecutor {
	return &retryExecutor{
		diContainer: diContainer,
		shard:       shard,
		stop:        make(chan struct{}),
		wg:          wg,
	}
}

func (ex *retryExecutor) Name() string {
	return "retryexecutor-" + ex.shard.GetShardId()
}

func (ex *retryExecutor) Start() {
	tw := util.NewTickWorker(ex.Name(), 1*time.Second, ex.stop, ex.handle, ex.wg)
	tw.Start()
}

func (ex *retryExecutor) Stop() {
	ex.stop <- struct{}{}
}

func (ex *retryExecutor) handle() {
	actions, err := ex.shard.PollRetry(10)
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions.Actions {
		flowMachine, err := flow.GetFlowStateMachine(action.WorkflowName, action.FlowId, ex.diContainer)
		if err != nil {
			logger.Error("error in executing workflow", zap.String("wfName", action.WorkflowName), zap.String("flowId", action.FlowId), zap.Error(err))
			continue
		}
		err = flowMachine.DispatchAction(int(action.ActionId), int(action.RetryCount)+1)
		if err != nil {
			logger.Error("error in retrying workflow", zap.String("wfName", action.WorkflowName), zap.String("flowId", action.FlowId), zap.Error(err))
		}
	}
}