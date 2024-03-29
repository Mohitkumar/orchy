package service

import (
	"time"

	"github.com/mohitkumar/orchy/analytics"
	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/cluster"
	"github.com/mohitkumar/orchy/metadata"
	"github.com/mohitkumar/orchy/util"
)

type ActionExecutionService struct {
	cluster         *cluster.Cluster
	metadataService metadata.MetadataService
}

func NewActionExecutionService(cluster *cluster.Cluster, metadataService metadata.MetadataService) *ActionExecutionService {
	return &ActionExecutionService{
		cluster:         cluster,
		metadataService: metadataService,
	}
}
func (ts *ActionExecutionService) Poll(actionName string, batchSize int) (*api.Actions, error) {
	actions, err := ts.cluster.Poll(actionName, batchSize)
	if err != nil {
		return nil, err
	}
	for _, action := range actions.Actions {
		actionDef, _ := ts.metadataService.GetMetadataStorage().GetActionDefinition(action.ActionName)
		if actionDef.RetryCount > 1 {
			ts.cluster.Timeout(action.WorkflowName, action.FlowId, action.ActionName, int(action.ActionId), time.Duration(actionDef.TimeoutSeconds)*time.Second)
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
	actionName := actionResult.ActionName
	data := util.ConvertFromProto(actionResult.Data)

	switch actionResult.Status {
	case api.ActionResult_SUCCESS:
		analytics.RecordActionSuccess(wfName, wfId, actionName, int(actionResult.ActionId), data)
		s.cluster.ExecuteAction(wfName, wfId, "default", int(actionResult.ActionId), data)
	case api.ActionResult_FAIL:
		analytics.RecordActionFailure(wfName, wfId, actionName, int(actionResult.ActionId), "worker failed")
		s.cluster.RetryAction(wfName, wfId, actionName, int(actionResult.ActionId), "failed")
	}
	return nil
}
