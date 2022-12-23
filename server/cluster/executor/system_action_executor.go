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

type systemActionExecutor struct {
	diContainer *container.DIContiner
	shard       persistence.Shard
	wg          *sync.WaitGroup
	stop        chan struct{}
}

func NewSystemActionExecutor(diContainer *container.DIContiner, shard persistence.Shard, wg *sync.WaitGroup) *systemActionExecutor {
	return &systemActionExecutor{
		diContainer: diContainer,
		shard:       shard,
		stop:        make(chan struct{}),
		wg:          wg,
	}
}

func (ex *systemActionExecutor) Name() string {
	return "system-action-executor-" + ex.shard.GetShardId()
}

func (ex *systemActionExecutor) Start() {
	tw := util.NewTickWorker(ex.Name(), 1*time.Second, ex.stop, ex.handle, ex.wg)
	tw.Start()
}

func (ex *systemActionExecutor) Stop() {
	ex.stop <- struct{}{}
}

func (ex *systemActionExecutor) handle() {
	actions, err := ex.shard.PollAction("system", 10)
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions.Actions {
		flowMachine, err := flow.GetFlowStateMachine(action.WorkflowName, action.FlowId, ex.diContainer)
		if err != nil {
			logger.Error("error in executing workflow", zap.String("wfName", action.WorkflowName), zap.String("flowId", action.FlowId), zap.Error(err))
			continue
		}
		err = flowMachine.ExecuteSystemAction(1, int(action.ActionId))
		if err != nil {
			logger.Error("error in executing workflow", zap.String("wfName", action.WorkflowName), zap.String("flowId", action.FlowId), zap.Error(err))
		}
	}
}
