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
	return ts.container.GetExternalQueue().Poll(actionName, batchSize)
}

func (ts *ActionExecutionService) Push(res *api.ActionResult) error {
	return ts.HandleActionResult(res)
}

func (s *ActionExecutionService) HandleActionResult(actionResult *api.ActionResult) error {
	wfName := actionResult.WorkflowName
	wfId := actionResult.FlowId
	data := util.ConvertFromProto(actionResult.Data)
	flowMachine, err := flow.GetFlowStateMachine(wfName, wfId, s.container)
	if err != nil {
		return err
	}
	switch actionResult.Status {
	case api.ActionResult_SUCCESS:
		completed, err := flowMachine.MoveForwardAndDispatch("default", data)
		if completed {
			return nil
		}
		if err != nil {
			logger.Error("error moving forward in workflow", zap.Error(err))
			return err
		}
	case api.ActionResult_FAIL:
		actionDefinition, err := s.container.GetMetadataStorage().GetActionDefinition(actionResult.ActionName)
		if err != nil {
			logger.Error("action definition not found ", zap.String("action", actionResult.ActionName), zap.Error(err))
			return err
		}
		if actionResult.RetryCount <= int32(actionDefinition.RetryCount) {
			var retryAfter time.Duration
			switch actionDefinition.RetryPolicy {
			case model.RETRY_POLICY_FIXED:
				retryAfter = time.Duration(actionDefinition.RetryAfterSeconds) * time.Second
			case model.RETRY_POLICY_BACKOFF:
				retryAfter = time.Duration(actionDefinition.RetryAfterSeconds*int(actionResult.RetryCount)) * time.Second
			}
			err = flowMachine.RetryAction(int(actionResult.ActionId), int(actionResult.RetryCount)+1, retryAfter)

			if err != nil {
				return err
			}
		} else {
			logger.Error("action max retry exhausted, failing workflow", zap.Int("maxRetry", actionDefinition.RetryCount))
			flowMachine.MarkFailed()
		}
	}
	return nil
}
