package flow

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/action"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

type FlowMachine struct {
	WorkflowName   string
	FlowId         string
	flow           *Flow
	flowContext    *model.FlowContext
	container      *container.DIContiner
	CurrentActions map[int]action.Action
	completed      bool
}

func NewFlowStateMachine(container *container.DIContiner) *FlowMachine {
	return &FlowMachine{
		container:      container,
		CurrentActions: make(map[int]action.Action),
	}
}

func GetFlowStateMachine(wfName string, flowId string, container *container.DIContiner) (*FlowMachine, error) {
	flowMachine := &FlowMachine{
		WorkflowName:   wfName,
		FlowId:         flowId,
		container:      container,
		CurrentActions: make(map[int]action.Action),
	}
	wf, _ := container.GetMetadataStorage().GetWorkflowDefinition(wfName)
	flowMachine.flow = Convert(wf, flowId, container)
	flowCtx, err := container.GetClusterStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return nil, err
	}
	flowMachine.flowContext = flowCtx
	for actionId := range flowMachine.flowContext.CurrentActionIds {
		flowMachine.CurrentActions[actionId] = flowMachine.flow.Actions[actionId]
	}
	if flowMachine.flowContext.State == model.COMPLETED || flowMachine.flowContext.State == model.FAILED {
		flowMachine.completed = true
	}
	return flowMachine, nil
}

func (f *FlowMachine) InitAndDispatchAction(wfName string, input map[string]any) error {
	wf, err := f.container.GetMetadataStorage().GetWorkflowDefinition(wfName)
	if err != nil {
		return fmt.Errorf("workflow %s not found", wfName)
	}
	flowId := uuid.New().String()

	f.WorkflowName = wfName
	f.flow = Convert(wf, flowId, f.container)
	f.CurrentActions[wf.RootAction] = f.flow.Actions[wf.RootAction]
	f.FlowId = flowId
	dataMap := make(map[string]any)
	dataMap["input"] = input
	f.flowContext = &model.FlowContext{
		Id:               flowId,
		State:            model.RUNNING,
		CurrentActionIds: map[int]bool{wf.RootAction: true},
		Data:             dataMap,
	}
	f.flowContext.State = model.RUNNING
	return f.saveContextAndDispatchAction([]int{wf.RootAction}, 1)
}

func (f *FlowMachine) MoveForwardAndDispatch(event string, actionId int, dataMap map[string]any) (bool, error) {
	currentActionId := f.CurrentActions[actionId].GetId()
	nextActionMap := f.CurrentActions[actionId].GetNext()
	if f.completed {
		return true, nil
	}

	if len(nextActionMap) == 0 {
		delete(f.CurrentActions, currentActionId)
		delete(f.flowContext.CurrentActionIds, currentActionId)
		if len(f.flowContext.CurrentActionIds) == 0 {
			f.MarkComplete()
			return true, nil
		} else {
			f.container.GetClusterStorage().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
			return false, nil
		}
	}
	delete(f.CurrentActions, currentActionId)
	delete(f.flowContext.CurrentActionIds, currentActionId)
	for _, actId := range nextActionMap[event] {
		f.CurrentActions[actId] = f.flow.Actions[actId]
		f.flowContext.CurrentActionIds[actId] = true
	}
	data := f.flowContext.Data
	if dataMap != nil || len(dataMap) > 0 {
		output := make(map[string]any)
		output["output"] = dataMap
		data[fmt.Sprintf("%d", currentActionId)] = output
	}
	f.flowContext.Data = data
	actionIds := make([]int, 0, len(f.CurrentActions))
	for k := range f.CurrentActions {
		actionIds = append(actionIds, k)
	}

	err := f.saveContextAndDispatchAction(actionIds, 1)
	if err != nil {
		return false, err
	}
	return false, nil
}
func (f *FlowMachine) DispatchAction(actionId int, tryCount int) error {
	err := f.ValidateExecutionRequest([]int{actionId})
	if err != nil {
		return err
	}
	currentAction := f.CurrentActions[actionId]
	var actionType api.Action_Type
	if currentAction.GetType() == action.ACTION_TYPE_SYSTEM {
		actionType = api.Action_SYSTEM
	} else {
		actionType = api.Action_USER
	}
	act := &api.Action{
		WorkflowName: f.WorkflowName,
		FlowId:       f.FlowId,
		Data:         util.ConvertToProto(util.ResolveInputParams(f.flowContext, currentAction.GetInputParams())),
		ActionId:     int32(currentAction.GetId()),
		ActionName:   currentAction.GetName(),
		RetryCount:   int32(tryCount),
		Type:         actionType,
	}

	err = f.container.GetClusterStorage().DispatchAction(f.FlowId, []*api.Action{act})
	if err != nil {
		return err
	}
	return nil
}

func (f *FlowMachine) saveContextAndDispatchAction(actionIds []int, tryCount int) error {
	err := f.ValidateExecutionRequest(actionIds)
	if err != nil {
		return err
	}
	var actions []*api.Action
	for _, actionId := range actionIds {
		currentAction := f.CurrentActions[actionId]
		var actionType api.Action_Type
		if currentAction.GetType() == action.ACTION_TYPE_SYSTEM {
			actionType = api.Action_SYSTEM
		} else {
			actionType = api.Action_USER
		}
		act := &api.Action{
			WorkflowName: f.WorkflowName,
			FlowId:       f.FlowId,
			Data:         util.ConvertToProto(util.ResolveInputParams(f.flowContext, currentAction.GetInputParams())),
			ActionId:     int32(actionId),
			ActionName:   currentAction.GetName(),
			RetryCount:   int32(tryCount),
			Type:         actionType,
		}
		actions = append(actions, act)
	}
	err = f.container.GetClusterStorage().SaveFlowContextAndDispatchAction(f.WorkflowName, f.FlowId, f.flowContext, actions)
	if err != nil {
		return err
	}
	return nil
}

func (f *FlowMachine) RetryAction(actionId int, tryCount int, retryAfter time.Duration) error {
	err := f.ValidateExecutionRequest([]int{actionId})
	if err != nil {
		return err
	}
	currentAction := f.CurrentActions[actionId]
	act := &api.Action{
		WorkflowName: f.WorkflowName,
		FlowId:       f.FlowId,
		Data:         util.ConvertToProto(util.ResolveInputParams(f.flowContext, currentAction.GetInputParams())),
		ActionId:     int32(currentAction.GetId()),
		ActionName:   currentAction.GetName(),
		RetryCount:   int32(tryCount),
	}
	err = f.container.GetClusterStorage().Retry(act, retryAfter)

	if err != nil {
		return err
	}
	return nil
}

func (f *FlowMachine) DelayAction(actionId int, tryCount int, delay time.Duration) error {
	err := f.ValidateExecutionRequest([]int{actionId})
	if err != nil {
		return err
	}
	currentAction := f.CurrentActions[actionId]
	act := &api.Action{
		WorkflowName: f.WorkflowName,
		FlowId:       f.FlowId,
		Data:         util.ConvertToProto(util.ResolveInputParams(f.flowContext, currentAction.GetInputParams())),
		ActionId:     int32(currentAction.GetId()),
		ActionName:   currentAction.GetName(),
		RetryCount:   int32(tryCount),
	}
	err = f.container.GetClusterStorage().Delay(act, delay)

	if err != nil {
		return err
	}
	return nil
}

func (f *FlowMachine) MarkComplete() {
	f.flowContext.State = model.COMPLETED
	f.completed = true
	f.container.GetClusterStorage().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	successHandler := f.container.GetStateHandler().GetHandler(f.flow.SuccessHandler)
	err := successHandler(f.WorkflowName, f.FlowId)
	if err != nil {
		logger.Error("error in running success handler", zap.Error(err))
	}
	logger.Info("workflow completed", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
}

func (f *FlowMachine) MarkFailed() {
	f.flowContext.State = model.FAILED
	f.completed = true
	f.container.GetClusterStorage().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	failureHandler := f.container.GetStateHandler().GetHandler(f.flow.FailureHandler)
	err := failureHandler(f.WorkflowName, f.FlowId)
	if err != nil {
		logger.Error("error in running failure handler", zap.Error(err))
	}
	logger.Info("workflow failed", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
}

func (f *FlowMachine) MarkWaitingDelay() {
	f.flowContext.State = model.WAITING_DELAY
	f.container.GetClusterStorage().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	logger.Info("workflow waiting delay", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
}

func (f *FlowMachine) MarkWaitingEvent() {
	f.flowContext.State = model.WAITING_EVENT
	f.container.GetClusterStorage().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	logger.Info("workflow waiting for event", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
}

func (f *FlowMachine) MarkPaused() error {
	f.flowContext.State = model.PAUSED
	err := f.container.GetClusterStorage().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	if err != nil {
		return err
	}
	logger.Info("workflow paused", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
	return nil
}

func (f *FlowMachine) MarkRunning() error {
	f.flowContext.State = model.RUNNING
	err := f.container.GetClusterStorage().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	if err != nil {
		return err
	}
	logger.Info("workflow running", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
	return nil
}

func (f *FlowMachine) GetFlowState() model.FlowState {
	return f.flowContext.State
}

func (f *FlowMachine) Resume() error {
	f.flowContext.State = model.RUNNING
	keys := make([]int, 0, len(f.CurrentActions))
	for k := range f.CurrentActions {
		keys = append(keys, k)
	}

	return f.saveContextAndDispatchAction(keys, 1)
}

func (f *FlowMachine) ExecuteSystemAction(tryCount int, actionId int) error {
	err := f.ValidateExecutionRequest([]int{actionId})
	if err != nil {
		return err
	}
	currentAction := f.CurrentActions[actionId]
	event, dataMap, err := currentAction.Execute(f.WorkflowName, f.flowContext, tryCount)
	if err != nil {
		return err
	}

	if currentAction.GetType() == action.ACTION_TYPE_SYSTEM {
		switch currentAction.GetName() {
		case "delay":
			f.MarkWaitingDelay()
			delay := currentAction.GetParams()["delay"].(time.Duration)
			f.DelayAction(currentAction.GetId(), 1, delay)
		case "wait":
			f.MarkWaitingEvent()
		default:
			completed, err := f.MoveForwardAndDispatch(event, actionId, dataMap)
			if completed {
				return nil
			}
			if err != nil {
				logger.Error("error moving forward in workflow", zap.Error(err))
				return err
			}
		}
	} else {
		return fmt.Errorf("should be system action")
	}
	return nil
}

func (f *FlowMachine) ValidateExecutionRequest(actionIds []int) error {
	if f.GetFlowState() == model.COMPLETED {
		return fmt.Errorf("can not run completed flow")
	}
	if f.GetFlowState() == model.FAILED {
		return fmt.Errorf("can not run failed flow")
	}
	if f.GetFlowState() == model.PAUSED {
		return fmt.Errorf("can not run paused flow")
	}
	for _, actionId := range actionIds {
		if _, ok := f.CurrentActions[actionId]; !ok {
			return fmt.Errorf("action %d already executed", actionId)
		}
	}
	return nil
}

func (f *FlowMachine) IsCompleted() bool {
	return f.completed
}
