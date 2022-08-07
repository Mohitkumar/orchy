package action

import (
	"strconv"

	"github.com/mohitkumar/orchy/server/model"
	"github.com/oliveagle/jsonpath"
)

var _ Action = new(switchAction)

type switchAction struct {
	baseAction
	expression string
	cases      map[string]int
}

func NewSwitchAction(expression string, bAction baseAction) *switchAction {
	return &switchAction{
		baseAction: bAction,
		expression: expression,
	}
}

func (d *switchAction) GetNext() map[string]int {
	return d.baseAction.nextMap
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
