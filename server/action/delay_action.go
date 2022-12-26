package action

import (
	"fmt"
	"time"

	"github.com/mohitkumar/orchy/server/model"
)

var _ Action = new(delayAction)

type delayAction struct {
	baseAction
	delay time.Duration
}

func NewDelayAction(delaySeconds int, bAction baseAction) *delayAction {
	act := &delayAction{
		baseAction: bAction,
		delay:      time.Duration(delaySeconds) * time.Second,
	}
	act.baseAction.params["delay"] = act.delay
	return act
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
	return "default", nil, nil
}
