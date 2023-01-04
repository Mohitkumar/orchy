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
	container   *container.DIContiner
	flowService *flow.FlowService
}

func NewActionExecutionService(container *container.DIContiner, flowService *flow.FlowService) *ActionExecutionService {
	return &ActionExecutionService{
		container:   container,
		flowService: flowService,
	}
}
func (ts *ActionExecutionService) Poll(actionName string, batchSize int) (*api.Actions, error) {
	actions, err := ts.container.GetExternalQueue().Poll(actionName, batchSize)
	if err != nil {
		return nil, err
	}
	for _, action := range actions.Actions {
		actionDef, _ := ts.container.GetMetadataStorage().GetActionDefinition(action.ActionName)
		if actionDef.RetryCount > 1 {
			ts.container.GetClusterStorage().Timeout(action, time.Duration(actionDef.TimeoutSeconds)*time.Second)
		}
	}
	return actions, nil
}

func (ts *ActionExecutionService) PollStream(actionName string) (<-chan *api.Action, error) {
	ch := make(chan *api.Action)
	outCh := make(chan *api.Action)
	ts.container.GetExternalQueue().PollStream(actionName, ch)
	go func() {
		for {
			act := <-ch
			actionDef, _ := ts.container.GetMetadataStorage().GetActionDefinition(act.ActionName)
			if actionDef.RetryCount > 1 {
				ts.container.GetClusterStorage().Timeout(act, time.Duration(actionDef.TimeoutSeconds)*time.Second)
			}
			outCh <- act
		}
	}()
	return outCh, nil
}

func (ts *ActionExecutionService) Push(res *api.ActionResult) error {
	return ts.HandleActionResult(res)
}

func (s *ActionExecutionService) HandleActionResult(actionResult *api.ActionResult) error {
	wfName := actionResult.WorkflowName
	wfId := actionResult.FlowId
	data := util.ConvertFromProto(actionResult.Data)

	switch actionResult.Status {
	case api.ActionResult_SUCCESS:
		s.flowService.ExecuteAction(wfName, wfId, "default", int(actionResult.ActionId), 1, data)
	case api.ActionResult_FAIL:
		actionDefinition, err := s.container.GetMetadataStorage().GetActionDefinition(actionResult.ActionName)
		if err != nil {
			logger.Error("action definition not found ", zap.String("action", actionResult.ActionName), zap.Error(err))
			return err
		}
		if actionResult.RetryCount < int32(actionDefinition.RetryCount) {
			var retryAfter time.Duration
			switch actionDefinition.RetryPolicy {
			case model.RETRY_POLICY_FIXED:
				retryAfter = time.Duration(actionDefinition.RetryAfterSeconds) * time.Second
			case model.RETRY_POLICY_BACKOFF:
				retryAfter = time.Duration(actionDefinition.RetryAfterSeconds*int(actionResult.RetryCount+1)) * time.Second
			}
			s.flowService.RetryAction(wfName, wfId, int(actionResult.ActionId), int(actionResult.RetryCount)+1, retryAfter, "failed")
		} else {
			logger.Error("action max retry exhausted, failing workflow", zap.Int("maxRetry", actionDefinition.RetryCount))
			s.flowService.MarkFailed(wfName, wfId)
		}
	}
	return nil
}
