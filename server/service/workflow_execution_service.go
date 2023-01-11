package service

import (
	"github.com/mohitkumar/orchy/server/engine"
	"github.com/mohitkumar/orchy/server/logger"
	"go.uber.org/zap"
)

type WorkflowExecutionService struct {
	flowEngine *engine.FlowEngine
}

func NewWorkflowExecutionService(flowEngine *engine.FlowEngine) *WorkflowExecutionService {
	return &WorkflowExecutionService{
		flowEngine: flowEngine,
	}
}

func (s *WorkflowExecutionService) StartFlow(name string, input map[string]any) (string, error) {
	flowId, err := s.flowEngine.Init(name, input)
	if err != nil {
		return "", err
	}
	logger.Info("started workflow", zap.String("workflow", name), zap.Any("input", input))
	return flowId, nil
}

func (s *WorkflowExecutionService) ResumeFlow(name string, flowId string) error {
	logger.Info("resuming workflow", zap.String("workflow", name), zap.String("id", flowId))
	s.flowEngine.ExecuteResume(name, flowId, "default")
	return nil
}

func (s *WorkflowExecutionService) PauseFlow(name string, flowId string) error {
	logger.Info("pausing workflow", zap.String("workflow", name), zap.String("id", flowId))
	s.flowEngine.MarkPaused(name, flowId)
	return nil
}

func (s *WorkflowExecutionService) ConsumeEvent(name string, flowId string, event string) error {
	s.flowEngine.ExecuteResume(name, flowId, event)
	return nil
}
