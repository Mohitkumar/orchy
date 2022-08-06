package service

import (
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/executor"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
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
		flowMachine := flow.GetFlowStateMachine(wfName, wfId, s.container)
		flowMachine.MoveForward("default", data)

		req := model.ActionExecutionRequest{
			WorkflowName: wfName,
			ActionId:     flowMachine.CurrentAction.GetId(),
			FlowId:       wfId,
			RetryCount:   1,
		}
		s.actionExecutor.Execute(req)
	case api.TaskResult_FAIL:
		taskDef, err := s.container.GetTaskDao().GetTask(taskResult.TaskName)
		if err != nil {
			logger.Error("task definition not found ", zap.String("taskName", taskResult.TaskName), zap.Error(err))
			return err
		}
		if taskResult.RetryCount <= int32(taskDef.RetryCount) {
			var retryAfter time.Duration
			switch taskDef.RetryPolicy {
			case model.RETRY_POLICY_FIXED:
				retryAfter = time.Duration(taskDef.RetryAfterSeconds) * time.Second
			case model.RETRY_POLICY_BACKOFF:
				retryAfter = time.Duration(taskDef.RetryAfterSeconds*int(taskResult.RetryCount)) * time.Second
			}
			req := model.ActionExecutionRequest{
				WorkflowName: wfName,
				ActionId:     int(taskResult.ActionId),
				FlowId:       wfId,
				RetryCount:   int(taskResult.RetryCount) + 1,
			}
			data, err := s.container.ActionExecutionRequestEncDec.Encode(req)
			if err != nil {
				return err
			}
			s.container.GetTaskRetryQueue().PushWithDelay("retry-queue", retryAfter, data)
		} else {
			logger.Error("task max retry exhausted, failing workflow", zap.Int("maxRetry", taskDef.RetryCount))
			flowCtx, err := s.container.GetFlowDao().GetFlowContext(wfName, wfId)
			if err != nil {
				return err
			}
			flowCtx.State = model.FAILED
			return s.container.GetFlowDao().SaveFlowContext(wfName, wfId, flowCtx)
		}
	}
	return nil
}
