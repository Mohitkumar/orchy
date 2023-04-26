package action

import (
	"encoding/json"
	"fmt"

	"github.com/dop251/goja"
	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/model"
	"go.uber.org/zap"
)

var _ Action = new(jsAction)

type jsAction struct {
	baseAction
	expression string
}

func NewJsAction(expression string, bAction baseAction) *jsAction {
	return &jsAction{
		baseAction: bAction,
		expression: expression,
	}
}

func (d *jsAction) Validate() error {
	if len(d.expression) == 0 {
		return fmt.Errorf("actionId=%d, expression can not be empty", d.id)
	}
	return nil
}

func (d *jsAction) GetNext() map[string][]int {
	return d.baseAction.nextMap
}

func (d *jsAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	logger.Info("running action", zap.String("name", d.name), zap.String("workflow", wfName), zap.String("id", flowContext.Id))
	data, _ := json.Marshal(flowContext.Data)
	expression := fmt.Sprintf("var $ = %s;\n", data)
	expression = expression + d.expression
	vm := goja.New()
	_, err := vm.RunString(expression)
	if err != nil {
		return "", nil, fmt.Errorf("error executing javascript %w", err)
	}
	val, err := vm.RunString("$")
	if err != nil {
		return "", nil, fmt.Errorf("error executing javascript %w", err)
	}
	res, err := json.Marshal(val.Export())
	if err != nil {
		return "", nil, err
	}
	var output map[string]any
	json.Unmarshal(res, &output)
	return "default", output, nil
}
