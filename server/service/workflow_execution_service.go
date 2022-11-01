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
func (s *WorkflowExecutionService) ResumeFlows() {
	flows, err := s.container.GetFlowDao().GetRunningFlows()
	if err != nil {
		logger.Error("can not get running flows", zap.Error(err))
	}
	for name, ids := range flows {
		for _, id := range ids {
			err := s.ResumeFlow(name, id)
			if err != nil {
				logger.Error("error resuming flow", zap.String("name", name), zap.String("id", id), zap.Error(err))
			}
		}
	}
}

func (s *WorkflowExecutionService) StartFlow(name string, input map[string]any) (string, error) {
	flowMachine := flow.NewFlowStateMachine(s.container)
	err := flowMachine.Init(name, input)
	if err != nil {
		return "", err
	}
	logger.Info("starting workflow", zap.String("workflow", name), zap.Any("input", input))
	return flowMachine.FlowId, flowMachine.Execute(1, flowMachine.CurrentAction.GetId())
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
	completed, err := flowMachine.MoveForward("default", nil)
	if completed {
		return nil
	}
	if err != nil {
		return err
	}
	flowMachine.MarkRunning()
	return flowMachine.Execute(1, flowMachine.CurrentAction.GetId())
}
