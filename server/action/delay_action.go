package action

import (
	"encoding/json"
	"time"

	"github.com/mohitkumar/orchy/server/model"
)

var _ Action = new(delayAction)

type delayAction struct {
	baseAction
	delay time.Duration
}

func NewDelayAction(delaySeconds int, bAction baseAction) *delayAction {
	return &delayAction{
		baseAction: bAction,
		delay:      time.Duration(delaySeconds) * time.Second,
	}
}

func (d *delayAction) GetNext() map[string]int {
	return d.baseAction.nextMap
}

func (d *delayAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	msg := model.ActionExecutionRequest{
		WorkflowName: wfName,
		FlowId:       flowContext.Id,
		ActionId:     d.nextMap["default"],
	}
	d.container.ActionExecutionRequestEncDec.Encode(msg)
	data, _ := json.Marshal(msg)
	err := d.container.GetDelayQueue().PushWithDelay("delay_action", d.delay, data)
	if err != nil {
		return "", nil, err
	}
	return "default", nil, nil
}
