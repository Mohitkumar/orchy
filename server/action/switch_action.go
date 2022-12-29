package action

import (
	"fmt"
	"strconv"

	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/oliveagle/jsonpath"
	"go.uber.org/zap"
)

var _ Action = new(switchAction)

type switchAction struct {
	baseAction
	expression string
}

func NewSwitchAction(expression string, bAction baseAction) *switchAction {
	return &switchAction{
		baseAction: bAction,
		expression: expression,
	}
}

func (d *switchAction) Validate() error {
	if len(d.expression) == 0 {
		return fmt.Errorf("actionId=%d, expression can not be empty", d.id)
	}
	_, err := jsonpath.Compile(d.expression)
	if err != nil {
		return fmt.Errorf("actionId=%d, expression should be a valid jsonpath expression", d.id)
	}
	if len(d.nextMap) == 0 {
		return fmt.Errorf("actionId=%d, switch action should have at least one  next action id", d.id)
	}
	return nil
}

func (d *switchAction) GetNext() map[string][]int {
	return d.baseAction.nextMap
}
func (d *switchAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	logger.Info("running action", zap.String("name", d.name), zap.String("workflow", wfName), zap.String("id", flowContext.Id))
	dataMap := flowContext.Data
	expressionValue, err := jsonpath.JsonPathLookup(dataMap, d.expression)
	event := "default"
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
