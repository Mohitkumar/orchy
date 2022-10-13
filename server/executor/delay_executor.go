package executor

import (
	"sync"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ Executor = new(DelayExecutor)

type DelayExecutor struct {
	container *container.DIContiner
	wg        *sync.WaitGroup
	stop      chan struct{}
}

func NewDelayExecutor(container *container.DIContiner, wg *sync.WaitGroup) *DelayExecutor {
	return &DelayExecutor{
		container: container,
		stop:      make(chan struct{}),
		wg:        wg,
	}
}

func (ex *DelayExecutor) Name() string {
	return "delay-executor"
}

func (ex *DelayExecutor) handle(msg *model.ActionExecutionRequest) {
	flowMachine, err := flow.GetFlowStateMachine(msg.WorkflowName, msg.FlowId, ex.container)
	if err != nil {
		return
	}
	completed, err := flowMachine.MoveForward("default", nil)
	if completed {
		return
	}
	if err != nil {
		logger.Error("error moving forward in workflow", zap.Error(err))
		return
	}

	flowMachine.MarkRunning()
	err = flowMachine.Execute(msg.TryNumber, msg.ActionId)
	if err != nil {
		logger.Error("error in executing workflow", zap.String("wfName", msg.WorkflowName), zap.String("flowId", msg.FlowId))
		return
	}
}
func (ex *DelayExecutor) Start() error {
	fn := func() {
		res, err := ex.container.GetDelayQueue().Pop("delay_action")
		if err != nil {
			logger.Error("error while polling delay queue", zap.Error(err))
			return
		}
		for _, r := range res {
			msg, err := ex.container.ActionExecutionRequestEncDec.Decode([]byte(r))
			if err != nil {
				logger.Error("can not decode action execution request")
				continue
			}
			ex.handle(msg)
		}
	}
	tw := util.NewTickWorker("delay-worker", 1, ex.stop, fn, ex.wg)
	tw.Start()
	logger.Info("delay executor started")
	return nil
}

func (ex *DelayExecutor) Stop() error {
	ex.stop <- struct{}{}
	return nil
}
