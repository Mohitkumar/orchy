package service

import (
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"go.uber.org/zap"
)

type WorkflowExecutionService struct {
	container *container.DIContiner
}

func NewWorkflowExecutionService(container *container.DIContiner) *WorkflowExecutionService {
	return &WorkflowExecutionService{
		container: container,
	}
}

func (s *WorkflowExecutionService) StartFlow(name string, input map[string]any) (string, error) {
	flowMachine := flow.NewFlowStateMachine(s.container)
	err := flowMachine.InitAndDispatchAction(name, input)
	if err != nil {
		return "", err
	}
	logger.Info("started workflow", zap.String("workflow", name), zap.Any("input", input))
	return flowMachine.FlowId, nil
}

func (s *WorkflowExecutionService) ResumeFlow(name string, flowId string) error {
	logger.Info("resuming workflow", zap.String("workflow", name), zap.String("id", flowId))
	flowMachine, err := flow.GetFlowStateMachine(name, flowId, s.container)
	if err != nil {
		return err
	}
	return flowMachine.Resume()
}

func (s *WorkflowExecutionService) PauseFlow(name string, flowId string) error {
	logger.Info("pausing workflow", zap.String("workflow", name), zap.String("id", flowId))
	flowMachine, err := flow.GetFlowStateMachine(name, flowId, s.container)
	if err != nil {
		return err
	}
	return flowMachine.MarkPaused()
}

func (s *WorkflowExecutionService) ConsumeEvent(name string, flowId string, event string) error {
	flowMachine, err := flow.GetFlowStateMachine(name, flowId, s.container)
	if err != nil {
		return err
	}
	completed, err := flowMachine.MoveForwardAndDispatch("default", nil, flowMachine.CurrentAction.GetId(), 1)
	if completed {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}
