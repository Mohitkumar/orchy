package service

import (
	"fmt"

	"github.com/google/uuid"
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
	wf, err := s.container.GetWorkflowDao().Get(name)
	if err != nil {
		logger.Error("workflow not found", zap.String("name", name))
		return fmt.Errorf("workflow = %s not found", name)
	}
	flow := flow.Convert(wf, uuid.New().String(), s.container)
	dataMap := make(map[string]any)
	dataMap["input"] = input
	flowCtx := &model.FlowContext{
		Id:            flow.Id,
		State:         model.RUNNING,
		CurrentAction: wf.RootAction,
		Data:          dataMap,
	}
	err = s.container.GetFlowDao().SaveFlowContext(name, flow.Id, flowCtx)
	if err != nil {
		return err
	}
	logger.Info("starting workflow", zap.String("workflow", name), zap.Int("rootAction", flow.RootAction))
	req := model.ActionExecutionRequest{
		WorkflowName: name,
		ActionId:     wf.RootAction,
		FlowId:       flow.Id,
		RetryCount:   1,
	}
	return s.actionExecutor.Execute(req)
}
