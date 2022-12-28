package action

import (
	"fmt"

	"github.com/mohitkumar/orchy/server/model"
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
	_, err := ua.container.GetMetadataStorage().GetActionDefinition(ua.name)
	if err != nil {
		return fmt.Errorf("actionId=%d, action %s not registered", ua.id, ua.name)
	}

	return nil
}
func (ua *UserAction) GetNext() map[string][]int {
	return ua.baseAction.nextMap
}

func (ua *UserAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	return "default", nil, nil
}
