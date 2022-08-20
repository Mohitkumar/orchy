package executor

import (
	"fmt"
	"sync"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ Executor = new(ActionExecutor)

type ActionExecutor struct {
	container *container.DIContiner
	capacity  int
	worker    *util.Worker
	wg        *sync.WaitGroup
}

func NewActionExecutor(container *container.DIContiner, capacity int, wg *sync.WaitGroup) *ActionExecutor {
	return &ActionExecutor{
		container: container,
		capacity:  capacity,
		wg:        wg,
	}
}

func (ex *ActionExecutor) handler(task util.Task) error {
	req, ok := task.(model.ActionExecutionRequest)
	if !ok {
		return fmt.Errorf("can not handle taks of type other than model.ActionExecutionRequest")
	}
	actionId := req.ActionId
	wfName := req.WorkflowName
	flowId := req.FlowId
	tryNumber := req.TryNumber
	flowMachine, err := flow.GetFlowStateMachine(wfName, flowId, ex.container)
	if err != nil {
		return err
	}
	if actionId != flowMachine.CurrentAction.GetId() {
		logger.Error("action already executed", zap.Int("actionIdRequested", actionId), zap.Int("flowCurrentActionId", flowMachine.CurrentAction.GetId()))
		return fmt.Errorf("action %d already executed", actionId)
	}
	err = flowMachine.Execute(tryNumber)
	if err != nil {
		return err
	}
	return nil
}

func (ex *ActionExecutor) Start() error {
	ex.worker = util.NewWorker("action-executor", ex.wg, ex.handler, ex.capacity)
	ex.worker.Start()
	logger.Info("action executor started")
	return nil
}

func (ex *ActionExecutor) Stop() error {
	ex.worker.Stop()
	return nil
}

func (ex *ActionExecutor) Name() string {
	return "action-executor"
}

func (ex *ActionExecutor) Execute(request model.ActionExecutionRequest) error {
	ex.worker.Sender() <- request
	return nil
}
