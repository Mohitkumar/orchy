package action

import (
	"encoding/json"
	"fmt"

	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"go.uber.org/zap"
	v8 "rogchap.com/v8go"
)

var _ Action = new(jsAction)

type jsAction struct {
	baseAction
	expression string
	jsvM       *v8.Isolate
}

func NewJsAction(expression string, bAction baseAction) *jsAction {
	return &jsAction{
		baseAction: bAction,
		expression: expression,
		jsvM:       v8.NewIsolate(),
	}
}

func (d *jsAction) Validate() error {
	if len(d.expression) < 0 {
		return fmt.Errorf("actionId=%d, expression can not be empty", d.id)
	}
	return nil
}

func (d *jsAction) GetNext() map[string]int {
	return d.baseAction.nextMap
}

func (d *jsAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	logger.Info("running action", zap.String("name", d.name), zap.String("workflow", wfName), zap.String("id", flowContext.Id))
	data, _ := json.Marshal(flowContext.Data)
	expression := fmt.Sprintf("var $ = %s;\n", data)
	expression = expression + d.expression
	fileName := fmt.Sprintf("%s_%s.js", wfName, flowContext.Id)
	ctx := v8.NewContext(d.jsvM)
	_, err := ctx.RunScript(expression, fileName)
	if err != nil {
		return "", nil, fmt.Errorf("error executing javascript %w", err)
	}
	val, err := ctx.RunScript("$", fileName)
	if err != nil {
		return "", nil, fmt.Errorf("error executing javascript %w", err)
	}
	js, err := val.MarshalJSON()
	if err != nil {
		return "", nil, err
	}
	var output map[string]any
	json.Unmarshal(js, &output)
	return "default", output, nil
}
