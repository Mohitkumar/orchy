package action

import (
	"fmt"
	"strings"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/oliveagle/jsonpath"
)

type ActionType string

const ACTION_TYPE_SYSTEM ActionType = "SYSTEM"
const ACTION_TYPE_USER ActionType = "USER"

func ToActionType(at string) ActionType {
	if strings.EqualFold(at, "system") {
		return ACTION_TYPE_SYSTEM
	}
	return ACTION_TYPE_USER
}

type Action interface {
	GetId() int
	GetName() string
	GetType() ActionType
	GetInputParams() map[string]any
	Execute(wfName string, flowContext *model.FlowContext) error
}

var _ Action = new(baseAction)

type baseAction struct {
	id          int
	actType     ActionType
	name        string
	inputParams map[string]any
	container   *container.DIContiner
}

func NewBaseAction(id int, Type ActionType, name string, inputParams map[string]any, container *container.DIContiner) *baseAction {
	return &baseAction{
		id:          id,
		name:        name,
		inputParams: inputParams,
		actType:     Type,
		container:   container,
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

func (ba *baseAction) Execute(wfName string, flowContext *model.FlowContext) error {
	return fmt.Errorf("can not execute")
}

func (ba *baseAction) ResolveInputParams(flowContext *model.FlowContext) map[string]any {
	flowData := flowContext.Data
	data := make(map[string]any)
	ba.resolveParams(flowData, ba.inputParams, data)
	return data
}

func (ba *baseAction) resolveParams(flowData map[string]any, params map[string]any, output map[string]any) {
	for k, v := range params {
		switch v.(type) {
		case map[string]any:
			out := make(map[string]any)
			output[k] = out
			ba.resolveParams(flowData, v.(map[string]any), out)
		case string:
			if strings.HasPrefix(v.(string), "$") {
				value, _ := jsonpath.JsonPathLookup(flowData, v.(string))
				output[k] = value
			} else {
				output[k] = v
			}
		default:
			output[k] = v
		}
	}
}
