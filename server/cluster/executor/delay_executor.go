package executor

import (
	"fmt"
	"sync"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ Executor = new(systemActionExecutor)

type delayExecutor struct {
	diContainer *container.DIContiner
	partition   int
	wg          *sync.WaitGroup
	stop        chan struct{}
}

func NewDelayExecutor(partition int, diContainer *container.DIContiner, wg *sync.WaitGroup) *delayExecutor {
	return &delayExecutor{
		diContainer: diContainer,
		partition:   partition,
		stop:        make(chan struct{}),
		wg:          wg,
	}
}

func (ex *delayExecutor) Name() string {
	return "delay-executor" + fmt.Sprintf("%d", ex.partition)
}

func (ex *delayExecutor) Start() {
	tw := util.NewTickWorker(ex.Name(), 1, ex.stop, ex.handle, ex.wg)
	tw.Start()
}

func (ex *delayExecutor) Stop() {
	ex.stop <- struct{}{}
}

func (ex *delayExecutor) handle() {
	actions, err := ex.diContainer.GetClusterStorage().PollDelay(10)
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions.Actions {
		flowMachine, err := flow.GetFlowStateMachine(action.WorkflowName, action.FlowId, ex.diContainer)
		if err != nil {
			logger.Error("error in executing workflow", zap.String("wfName", action.WorkflowName), zap.String("flowId", action.FlowId), zap.Error(err))
			continue
		}
		completed, err := flowMachine.MoveForwardAndDispatch("default", nil, int(action.ActionId), 1)
		if completed {
			continue
		}

		if err != nil {
			logger.Error("error in executing workflow", zap.String("wfName", action.WorkflowName), zap.String("flowId", action.FlowId), zap.Error(err))
		}
	}
}
