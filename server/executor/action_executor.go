package executor

import (
	"fmt"
	"sync"

	"github.com/mohitkumar/orchy/server/action"
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

func (ex *ActionExecutor) Start() error {
	handler := func(task util.Task) error {
		req, ok := task.(model.ActionExecutionRequest)
		if !ok {
			return fmt.Errorf("can not handle taks of type other than model.ActionExecutionRequest")
		}
		actionId := req.ActionId
		wfName := req.WorkflowName
		flowId := req.FlowId
		wf, err := ex.container.GetWorkflowDao().Get(wfName)
		if err != nil {
			logger.Error("workflow not found", zap.String("name", wfName))
			return fmt.Errorf("workflow = %s not found", wfName)
		}
		flow := flow.Convert(wf, flowId, ex.container)

		flowCtx, err := ex.container.GetFlowDao().GetFlowContext(wfName, flowId)
		if err != nil {
			return err
		}
		if _, ok := flow.Actions[int(actionId)]; !ok {
			flowCtx.State = model.COMPLETED
			return ex.container.GetFlowDao().SaveFlowContext(wfName, flowId, flowCtx)
		}
		currentAction := flow.Actions[actionId]
		err = currentAction.Execute(wfName, flowCtx)
		if err != nil {
			return err
		}
		nextActionId := flowCtx.NextAction

		switch currentAction.GetType() {
		case action.ACTION_TYPE_SYSTEM:
			switch currentAction.GetName() {
			case "switch":
				actionExReq := &model.ActionExecutionRequest{
					WorkflowName: wfName,
					ActionId:     nextActionId,
					FlowId:       flowId,
				}
				return ex.Execute(*actionExReq)
			case "delay":

			}
		case action.ACTION_TYPE_USER:

		}
		return nil
	}
	ex.worker = util.NewWorker("action-executor", ex.wg, handler, ex.capacity)
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
