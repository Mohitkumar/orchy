package action

import (
	"fmt"
	"time"

	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"go.uber.org/zap"
)

var _ Action = new(delayAction)

type waitAction struct {
	baseAction
	event   string
	timeout time.Duration
}

func NewWaitAction(eventName string, timeoutSeconds int, bAction baseAction) *waitAction {
	act := &waitAction{
		baseAction: bAction,
		event:      eventName,
		timeout:    time.Duration(timeoutSeconds) * time.Second,
	}
	act.baseAction.params["timeout"] = act.timeout
	act.baseAction.params["event"] = act.event
	return act
}

func (w *waitAction) GetNext() map[string][]int {
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
	logger.Info("running action", zap.String("name", d.name), zap.String("workflow", wfName), zap.String("id", flowContext.Id))
	return "default", nil, nil
}
