package action

import (
	"encoding/json"
	"time"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/model"
)

var _ Action = new(delayAction)

type delayAction struct {
	baseAction
	delay      time.Duration
	nextAction int
}

func NewDelayAction(id int, Type ActionType, name string, delaySeconds int, nextAction int, container *container.DIContiner) *delayAction {
	inputParams := map[string]any{}
	return &delayAction{
		baseAction: *NewBaseAction(id, Type, name, inputParams, container),
		delay:      time.Duration(delaySeconds) * time.Second,
		nextAction: nextAction,
	}
}

func (d *delayAction) GetNext() map[string]int {
	nextMap := make(map[string]int)
	nextMap["default"] = d.nextAction
	return nextMap
}

func (d *delayAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	msg := model.ActionExecutionRequest{
		WorkflowName: wfName,
		FlowId:       flowContext.Id,
		ActionId:     d.nextAction,
	}
	d.container.ActionExecutionRequestEncDec.Encode(msg)
	data, _ := json.Marshal(msg)
	err := d.container.GetDelayQueue().PushWithDelay("delay_action", d.delay, data)
	if err != nil {
		return "", nil, err
	}
	flowContext.State = model.WAITING_DELAY
	return "default", nil, d.container.GetFlowDao().SaveFlowContext(wfName, flowContext.Id, flowContext)
}
