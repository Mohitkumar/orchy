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

func (d *switchAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) error {
	dataMap := flowContext.Data
	expressionValue, err := jsonpath.JsonPathLookup(dataMap, d.expression)
	if err != nil {
		return err
	}
	var nextAction int
	switch expValue := expressionValue.(type) {
	case int, int16, int32, int64:
		nextAction = d.cases[strconv.Itoa(expressionValue.(int))]
	case float32, float64:
		nextAction = d.cases[strconv.Itoa(int(expressionValue.(float64)))]
	case bool:
		nextAction = d.cases[strconv.FormatBool(expressionValue.(bool))]
	case string:
		nextAction = d.cases[expValue]
	}
	flowContext.NextAction = nextAction
	err = d.container.GetFlowDao().SaveFlowContext(wfName, flowContext.Id, flowContext)
	if err != nil {
		return err
	}
	return nil
}
