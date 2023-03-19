package action

import (
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ Action = new(jsonMapAction)

type jsonMapAction struct {
	baseAction
}

func NewJsonMapAction(bAction baseAction) *jsonMapAction {
	return &jsonMapAction{
		baseAction: bAction,
	}
}

func (d *jsonMapAction) Validate() error {
	return nil
}

func (d *jsonMapAction) GetNext() map[string][]int {
	return d.baseAction.nextMap
}

func (d *jsonMapAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	logger.Info("running action", zap.String("name", d.name), zap.String("workflow", wfName), zap.String("id", flowContext.Id))
	output := util.ResolveInputParams(flowContext, d.inputParams)
	return "default", output, nil
}
