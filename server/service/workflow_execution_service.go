package service

import (
	"github.com/mohitkumar/orchy/server/cluster"
	"github.com/mohitkumar/orchy/server/logger"
	"go.uber.org/zap"
)

type WorkflowExecutionService struct {
	cluster *cluster.Cluster
}

func NewWorkflowExecutionService(cluster *cluster.Cluster) *WorkflowExecutionService {
	return &WorkflowExecutionService{
		cluster: cluster,
	}
}

func (s *WorkflowExecutionService) StartFlow(name string, input map[string]any) (string, error) {
	flowId, err := s.cluster.Init(name, input)
	if err != nil {
		return "", err
	}
	logger.Info("started workflow", zap.String("workflow", name), zap.Any("input", input))
	return flowId, nil
}

func (s *WorkflowExecutionService) ResumeFlow(name string, flowId string) error {
	logger.Info("resuming workflow", zap.String("workflow", name), zap.String("id", flowId))
	s.cluster.ExecuteResume(name, flowId, "default")
	return nil
}

func (s *WorkflowExecutionService) PauseFlow(name string, flowId string) error {
	logger.Info("pausing workflow", zap.String("workflow", name), zap.String("id", flowId))
	s.cluster.MarkPaused(name, flowId)
	return nil
}

func (s *WorkflowExecutionService) ConsumeEvent(name string, flowId string, event string) error {
	s.cluster.ExecuteResume(name, flowId, event)
	return nil
}
