package action

import (
	"fmt"
	"strings"

	"github.com/mohitkumar/orchy/model"
)

type ActionType string

const ACTION_TYPE_SYSTEM ActionType = "SYSTEM"
const ACTION_TYPE_USER ActionType = "USER"

var VALID_SYSTEM_ACTIONS = []string{"switch", "delay"}

func ToActionType(at string) ActionType {
	if strings.EqualFold(at, "system") {
		return ACTION_TYPE_SYSTEM
	}
	return ACTION_TYPE_USER
}

func ValidateActionType(at string) error {
	if strings.EqualFold(at, "system") || strings.EqualFold(at, "user") {
		return nil
	}
	return fmt.Errorf("invalid action type %s", at)
}

type Action interface {
	GetId() int
	GetName() string
	GetType() ActionType
	GetInputParams() map[string]any
	GetParams() map[string]any
	GetNext() map[string][]int
	Validate() error
	Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error)
}

var _ Action = new(baseAction)

type baseAction struct {
	id          int
	actType     ActionType
	name        string
	inputParams map[string]any
	nextMap     map[string][]int
	params      map[string]any
}

func NewBaseAction(id int, Type ActionType, name string, inputParams map[string]any, nextMap map[string][]int) *baseAction {
	return &baseAction{
		id:          id,
		name:        name,
		inputParams: inputParams,
		params:      make(map[string]any),
		actType:     Type,
		nextMap:     nextMap,
	}

}
func (ba *baseAction) GetId() int {
	return ba.id
}
func (ba *baseAction) GetName() string {
	return ba.name
}
func (ba *baseAction) GetType() ActionType {
	return ba.actType
}
func (ba *baseAction) GetInputParams() map[string]any {
	return ba.inputParams
}

func (ba *baseAction) GetParams() map[string]any {
	return ba.params
}

func (ba *baseAction) GetNext() map[string][]int {
	return ba.nextMap

}
func (ba *baseAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	return "default", nil, fmt.Errorf("can not execute")
}

func (ba *baseAction) Validate() error {
	return fmt.Errorf("action %s implementation not found", ba.name)
}
