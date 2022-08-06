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
func (s *WorkflowExecutionService) StartFlow(name string, input map[string]any) error {
	flowMachine := flow.NewFlowStateMachine(s.container)
	flowMachine.Init(name, input)
	logger.Info("starting workflow", zap.String("workflow", name))
	req := model.ActionExecutionRequest{
		WorkflowName: name,
		ActionId:     flowMachine.CurrentAction.GetId(),
		FlowId:       flowMachine.FlowId,
		RetryCount:   1,
	}
	return s.actionExecutor.Execute(req)
}
