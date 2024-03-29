package service

import (
	"github.com/mohitkumar/orchy/cluster"
	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/model"
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

func (s *WorkflowExecutionService) GetFlow(name string, flowId string) (*model.FlowContext, error) {
	return s.cluster.GetFlow(name, flowId)
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
	return s.cluster.ExecuteResumeAfterWait(name, flowId, event)
}
