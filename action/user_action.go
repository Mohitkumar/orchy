package action

import (
	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/model"
	"go.uber.org/zap"
)

var _ Action = new(UserAction)

type UserAction struct {
	baseAction
}

func NewUserAction(bAction baseAction) *UserAction {
	return &UserAction{
		baseAction: bAction,
	}
}

func (ua *UserAction) Validate() error {
	return nil
}
func (ua *UserAction) GetNext() map[string][]int {
	return ua.baseAction.nextMap
}

func (ua *UserAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	logger.Info("running action", zap.String("name", ua.name), zap.String("workflow", wfName), zap.String("id", flowContext.Id))
	return "default", nil, nil
}
