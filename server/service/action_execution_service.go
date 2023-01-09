package service

import (
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/cluster"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/util"
)

type ActionExecutionService struct {
	cluster   *cluster.Cluster
	flowService *flow.FlowService
}

func NewActionExecutionService(cluster *cluster.Cluster, flowService *flow.FlowService) *ActionExecutionService {
	return &ActionExecutionService{
		cluster:   cluster,
		flowService: flowService,
	}
}
func (ts *ActionExecutionService) Poll(actionName string, batchSize int) (*api.Actions, error) {
	actions, err := ts.cluster.().Poll(actionName, batchSize)
	if err != nil {
		return nil, err
	}
	for _, action := range actions.Actions {
		actionDef, _ := ts.cluster.GetMetadataStorage().GetActionDefinition(action.ActionName)
		if actionDef.RetryCount > 1 {
			ts.container.GetClusterStorage().Timeout(action, time.Duration(actionDef.TimeoutSeconds)*time.Second)
		}
	}
	return actions, nil
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
		s.flowService.ExecuteAction(wfName, wfId, "default", int(actionResult.ActionId), data)
	case api.ActionResult_FAIL:
		s.flowService.RetryAction(wfName, wfId, int(actionResult.ActionId), "failed")
	}
	return nil
}
