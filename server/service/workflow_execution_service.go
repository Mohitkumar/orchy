package service

import (
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"go.uber.org/zap"
)

type WorkflowExecutionService struct {
	flowService *flow.FlowService
}

func NewWorkflowExecutionService(flowService *flow.FlowService) *WorkflowExecutionService {
	return &WorkflowExecutionService{
		flowService: flowService,
	}
}

func (s *WorkflowExecutionService) StartFlow(name string, input map[string]any) (string, error) {
	flowId, err := s.flowService.Init(name, input)
	if err != nil {
		return "", err
	}
	logger.Info("started workflow", zap.String("workflow", name), zap.Any("input", input))
	return flowId, nil
}

func (s *WorkflowExecutionService) ResumeFlow(name string, flowId string) error {
	logger.Info("resuming workflow", zap.String("workflow", name), zap.String("id", flowId))
	return nil
}

func (s *WorkflowExecutionService) PauseFlow(name string, flowId string) error {
	logger.Info("pausing workflow", zap.String("workflow", name), zap.String("id", flowId))
	return nil
}

func (s *WorkflowExecutionService) ConsumeEvent(name string, flowId string, event string) error {
	s.flowService.ExecuteResume(name, flowId, event)
	return nil
}
