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

type delayExecutor struct {
	diContainer *container.DIContiner
	shard       persistence.Shard
	wg          *sync.WaitGroup
	tw          *util.TickWorker
	stop        chan struct{}
}

func NewDelayExecutor(diContainer *container.DIContiner, shard persistence.Shard, wg *sync.WaitGroup) *delayExecutor {
	ex := &delayExecutor{
		diContainer: diContainer,
		shard:       shard,
		stop:        make(chan struct{}),
		wg:          wg,
	}
	ex.tw = util.NewTickWorker(ex.Name(), 1*time.Second, ex.stop, ex.handle, ex.wg)
	return ex
}

func (ex *delayExecutor) Name() string {
	return "delay-executor-" + ex.shard.GetShardId()
}

func (ex *delayExecutor) Start() {
	if ex.IsRunning() {
		return
	}
	ex.tw.Start()
}

func (ex *delayExecutor) IsRunning() bool {
	return ex.tw.IsRunning()
}

func (ex *delayExecutor) Stop() {
	if !ex.IsRunning() {
		return
	}
	ex.stop <- struct{}{}
}

func (ex *delayExecutor) handle() {
	actions, err := ex.shard.PollDelay()
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions.Actions {
		flowMachine, err := flow.GetFlowStateMachine(action.WorkflowName, action.FlowId, ex.diContainer)
		if err != nil {
			logger.Error("error in executing workflow", zap.String("wfName", action.WorkflowName), zap.String("flowId", action.FlowId), zap.Error(err))
			continue
		}
		completed, err := flowMachine.MoveForwardAndDispatch("default", int(action.ActionId), nil)
		if completed {
			continue
		}

		if err != nil {
			logger.Error("error in executing workflow", zap.String("wfName", action.WorkflowName), zap.String("flowId", action.FlowId), zap.Error(err))
		}
	}
}
