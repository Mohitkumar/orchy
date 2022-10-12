package service

import (
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/executor"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"go.uber.org/zap"
)

type WorkflowExecutionService struct {
	container      *container.DIContiner
	actionExecutor *executor.ActionExecutor
}

func NewWorkflowExecutionService(container *container.DIContiner, actionExecutor *executor.ActionExecutor) *WorkflowExecutionService {
	return &WorkflowExecutionService{
		container:      container,
		actionExecutor: actionExecutor,
	}
}
func (s *WorkflowExecutionService) StartFlow(name string, input map[string]any) (string, error) {
	flowMachine := flow.NewFlowStateMachine(s.container)
	err := flowMachine.Init(name, input)
	if err != nil {
		return "", err
	}
	logger.Info("starting workflow", zap.String("workflow", name), zap.Any("input", input))
	req := model.ActionExecutionRequest{
		WorkflowName: name,
		ActionId:     flowMachine.CurrentAction.GetId(),
		FlowId:       flowMachine.FlowId,
		TryNumber:    1,
	}
	return flowMachine.FlowId, s.actionExecutor.Execute(req)
}

func (s *WorkflowExecutionService) ConsumeEvent(name string, flowId string, event string) error {
	flowMachine, err := flow.GetFlowStateMachine(name, flowId, s.container)
	if err != nil {
		return err
	}
	err = flowMachine.MoveForward("default", nil)
	if err != nil {
		return err
	}
	flowMachine.MarkRunning()
	return flowMachine.Execute(1)
}
