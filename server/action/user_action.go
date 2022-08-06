package action

import (
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var _ Action = new(UserAction)

type UserAction struct {
	baseAction
	nextAction int
}

func NewUserAction(id int, Type ActionType, name string, inputParams map[string]any, nextAction int, container *container.DIContiner) *UserAction {
	return &UserAction{
		baseAction: *NewBaseAction(id, Type, name, inputParams, container),
		nextAction: nextAction,
	}
}

func (ua *UserAction) GetNext() map[string]int {
	nextMap := make(map[string]int)
	nextMap["default"] = ua.nextAction
	return nextMap
}

func (ua *UserAction) Execute(wfName string, flowContext *model.FlowContext, retryCount int) (string, map[string]any, error) {
	taskDef, err := ua.container.GetTaskDao().GetTask(ua.name)
	if err != nil {
		logger.Error("task definition not found", zap.String("taskName", ua.name))
		return "", nil, err
	}
	task := &api.Task{
		WorkflowName: wfName,
		FlowId:       flowContext.Id,
		Data:         util.ConvertToProto(ua.ResolveInputParams(flowContext)),
		ActionId:     int32(flowContext.CurrentAction),
		TaskName:     ua.name,
		RetryCount:   int32(retryCount),
	}
	d, err := proto.Marshal(task)
	if err != nil {
		return "", nil, err
	}
	err = ua.container.GetQueue().Push(ua.GetName(), d)
	if err != nil {
		return "", nil, err
	}
	req := model.ActionExecutionRequest{
		WorkflowName: wfName,
		ActionId:     flowContext.CurrentAction,
		FlowId:       flowContext.Id,
		RetryCount:   retryCount,
		TaskName:     ua.name,
	}
	data, _ := ua.container.ActionExecutionRequestEncDec.Encode(req)
	ua.container.GetTaskTimeoutQueue().PushWithDelay("timeout-queue", time.Duration(taskDef.TimeoutSeconds)*time.Second, data)
	return "default", nil, nil
}
