package executor

import (
	"sync"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

type SystemActionExecutor struct {
	container *container.DIContiner
	wg        *sync.WaitGroup
	stop      chan struct{}
}

func NewSystemActionExecutor(container *container.DIContiner, wg *sync.WaitGroup) *SystemActionExecutor {
	return &SystemActionExecutor{
		container: container,
		stop:      make(chan struct{}),
		wg:        wg,
	}
}

func (ex *SystemActionExecutor) Name() string {
	return "system-executor"
}

func (ex *SystemActionExecutor) Start() error {
	fn := func() {
		res, err := ex.container.GetQueue().Pop("system", 1)
		if err != nil {
			logger.Error("error while polling retry queue", zap.Error(err))
			return
		}
		for _, r := range res {
			msg, err := ex.container.ActionExecutionRequestEncDec.Decode([]byte(r))
			if err != nil {
				logger.Error("can not decode action execution request")
				continue
			}
			flowMachine, err := flow.GetFlowStateMachine(msg.WorkflowName, msg.FlowId, ex.container)
			if err != nil {
				logger.Error("error in executing workflow", zap.String("wfName", msg.WorkflowName), zap.String("flowId", msg.FlowId))
				continue
			}
			err = flowMachine.ExecuteSystemAction(msg.TryNumber, msg.ActionId)
			if err != nil {
				logger.Error("error in executing workflow", zap.String("wfName", msg.WorkflowName), zap.String("flowId", msg.FlowId))
			}
		}
	}
	tw := util.NewTickWorker("retry-worker", 1, ex.stop, fn, ex.wg)
	tw.Start()
	logger.Info("system action executor started")
	return nil
}

func (ex *SystemActionExecutor) Stop() error {
	ex.stop <- struct{}{}
	return nil
}
