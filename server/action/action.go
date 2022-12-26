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
	GetNext() map[string]int
	Validate() error
	Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error)
}

var _ Action = new(baseAction)

type baseAction struct {
	id          int
	actType     ActionType
	name        string
	inputParams map[string]any
	nextMap     map[string]int
	params      map[string]any
	container   *container.DIContiner
}

func NewBaseAction(id int, Type ActionType, name string, inputParams map[string]any, nextMap map[string]int, container *container.DIContiner) *baseAction {
	return &baseAction{
		id:          id,
		name:        name,
		inputParams: inputParams,
		params:      make(map[string]any),
		actType:     Type,
		nextMap:     nextMap,
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

func (ba *baseAction) GetParams() map[string]any {
	return ba.params
}

func (ba *baseAction) GetNext() map[string]int {
	return ba.nextMap

}
func (ba *baseAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	return "default", nil, fmt.Errorf("can not execute")
}

func (ba *baseAction) Validate() error {
	return fmt.Errorf("action %s implementation not found", ba.name)
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
		case []any:
			l := v.([]any)
			output[k] = ba.resolveList(flowData, l)
		default:
			output[k] = v
		}
	}
}

func (ba *baseAction) resolveList(flowData map[string]any, list []any) []any {
	var output []any
	for _, v := range list {
		switch v.(type) {
		case map[string]any:
			out := make(map[string]any)
			output = append(output, out)
			ba.resolveParams(flowData, v.(map[string]any), out)
		case string:
			if strings.HasPrefix(v.(string), "$") {
				value, _ := jsonpath.JsonPathLookup(flowData, v.(string))
				output = append(output, value)
			} else {
				output = append(output, v)
			}
		case []any:
			l := v.([]any)
			outList := ba.resolveList(flowData, l)
			output = append(output, outList...)
		default:
			output = append(output, v)
		}
	}
	return output
}
