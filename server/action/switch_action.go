package action

import (
	"strconv"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/oliveagle/jsonpath"
)

var _ Action = new(switchAction)

type switchAction struct {
	baseAction
	expression string
	cases      map[string]int
}

func NewSwitchAction(id int, Type ActionType, name string, expression string, cases map[string]int, container *container.DIContiner) *switchAction {
	inputParams := map[string]any{}
	return &switchAction{
		baseAction: *NewBaseAction(id, Type, name, inputParams, container),
		expression: expression,
		cases:      cases,
	}
}

func (d *switchAction) GetNext() map[string]int {
	return d.cases
}
func (d *switchAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	dataMap := flowContext.Data
	expressionValue, err := jsonpath.JsonPathLookup(dataMap, d.expression)
	event := ""
	if err != nil {
		return event, nil, err
	}
	switch expValue := expressionValue.(type) {
	case int, int16, int32, int64:
		event = strconv.Itoa(expressionValue.(int))
	case float32, float64:
		event = strconv.Itoa(int(expressionValue.(float64)))
	case bool:
		event = strconv.FormatBool(expressionValue.(bool))
	case string:
		event = expValue
	}
	return event, nil, nil
}
