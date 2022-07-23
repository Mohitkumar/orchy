package service

import (
	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/executor"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/util"
	"google.golang.org/protobuf/proto"
)

type ActionExecutionService struct {
	container      *container.DIContiner
	actionExecutor *executor.ActionExecutor
}

func NewActionExecutionService(container *container.DIContiner, actionExecutor *executor.ActionExecutor) *ActionExecutionService {
	return &ActionExecutionService{
		container:      container,
		actionExecutor: actionExecutor,
	}
}
func (ts *ActionExecutionService) Poll(taskName string) (*api.Task, error) {
	data, err := ts.container.GetQueue().Pop(taskName)
	if err != nil {
		return nil, err
	}
	task := &api.Task{}
	err = proto.Unmarshal(data, task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (ts *ActionExecutionService) Push(res *api.TaskResult) error {
	return ts.HandleTaskResult(res)
}

func (s *ActionExecutionService) HandleTaskResult(taskResult *api.TaskResult) error {
	wfName := taskResult.WorkflowName
	wfId := taskResult.FlowId
	data := util.ConvertFromProto(taskResult.Data)
	switch taskResult.Status {
	case api.TaskResult_SUCCESS:

		flowCtx, err := s.container.GetFlowDao().AddActionOutputToFlowContext(wfName, wfId, int(taskResult.ActionId), data)
		if err != nil {
			return err
		}
		req := model.ActionExecutionRequest{
			WorkflowName: wfName,
			ActionId:     int(flowCtx.NextAction),
			FlowId:       wfId,
		}
		s.actionExecutor.Execute(req)
	case api.TaskResult_FAIL:
		//retry logic
	}
	return nil
}
