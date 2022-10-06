package action

import (
	"fmt"
	"time"

	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"go.uber.org/zap"
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

func (d *delayAction) Validate() error {
	if d.delay <= 0 {
		return fmt.Errorf("actionId=%d, delay value %d wrong", d.id, d.delay)
	}
	if _, ok := d.nextMap["default"]; !ok {
		return fmt.Errorf("actionId=%d, Delay action should have default next action id", d.id)
	}
	return nil
}

func (d *delayAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	logger.Info("running action", zap.String("name", d.name), zap.String("workflow", wfName), zap.String("id", flowContext.Id))
	msg := model.ActionExecutionRequest{
		WorkflowName: wfName,
		FlowId:       flowContext.Id,
		ActionId:     d.nextMap["default"],
	}
	data, err := d.container.ActionExecutionRequestEncDec.Encode(msg)
	err = d.container.GetDelayQueue().PushWithDelay("delay_action", flowContext.Id, d.delay, data)
	if err != nil {
		return "", nil, err
	}
	return "default", nil, nil
}
