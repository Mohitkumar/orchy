package action

import (
	"fmt"

	"github.com/mohitkumar/orchy/server/model"
)

var _ Action = new(delayAction)

type waitAction struct {
	baseAction
	event string
}

func NewWaitAction(eventName string, bAction baseAction) *waitAction {
	return &waitAction{
		baseAction: bAction,
		event:      eventName,
	}
}

func (w *waitAction) GetNext() map[string]int {
	return w.baseAction.nextMap
}

func (w *waitAction) Validate() error {
	if len(w.event) == 0 {
		return fmt.Errorf("actionId=%d, wait action should have event", w.id)
	}
	if _, ok := w.nextMap["default"]; !ok {
		return fmt.Errorf("actionId=%d, Wait action should have default next action id", w.id)
	}
	return nil
}

func (d *waitAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	return "default", nil, nil
}
