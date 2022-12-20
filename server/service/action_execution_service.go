package service

import (
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ActionExecutionService struct {
	container *container.DIContiner
}

func NewActionExecutionService(container *container.DIContiner) *ActionExecutionService {
	return &ActionExecutionService{
		container: container,
	}
}
func (ts *ActionExecutionService) Poll(actionName string, batchSize int) (*api.Actions, error) {
	msgs, err := ts.container.GetClusterStorage().PollAction(actionName, batchSize)
	if err != nil {
		return nil, err
	}
	taskArray := make([]*api.Task, 0)
	for _, msg := range msgs {
		task := &api.Task{}
		err = proto.Unmarshal([]byte(msg), task)
		if err != nil {
			continue
		}
		taskArray = append(taskArray, task)
	}

	tasks := &api.Tasks{
		Tasks: taskArray,
	}
	return tasks, nil
}

func (ts *ActionExecutionService) Push(res *api.TaskResult) error {
	return ts.HandleTaskResult(res)
}

func (s *ActionExecutionService) HandleTaskResult(taskResult *api.TaskResult) error {
	wfName := taskResult.WorkflowName
	wfId := taskResult.FlowId
	data := util.ConvertFromProto(taskResult.Data)
	flowMachine, err := flow.GetFlowStateMachine(wfName, wfId, s.container)
	if err != nil {
		return err
	}
	switch taskResult.Status {
	case api.TaskResult_SUCCESS:
		completed, err := flowMachine.MoveForward("default", data)
		if completed {
			return nil
		}
		if err != nil {
			logger.Error("error moving forward in workflow", zap.Error(err))
			return err
		}
		return flowMachine.Execute(1, flowMachine.CurrentAction.GetId())
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
				TryNumber:    int(taskResult.RetryCount) + 1,
			}
			data, err := s.container.ActionExecutionRequestEncDec.Encode(req)
			if err != nil {
				return err
			}
			s.container.GetTaskRetryQueue().PushWithDelay("retry-queue", wfId, retryAfter, data)
		} else {
			logger.Error("task max retry exhausted, failing workflow", zap.Int("maxRetry", taskDef.RetryCount))
			flowMachine.MarkFailed()
		}
	}
	return nil
}
