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
	WorkflowName  string
	FlowId        string
	flow          *Flow
	flowContext   *model.FlowContext
	container     *container.DIContiner
	CurrentAction action.Action
	completed     bool
}

func NewFlowStateMachine(container *container.DIContiner) *FlowMachine {
	return &FlowMachine{
		container: container,
	}
}

func GetFlowStateMachine(wfName string, flowId string, container *container.DIContiner) (*FlowMachine, error) {
	flowMachine := &FlowMachine{
		WorkflowName: wfName,
		FlowId:       flowId,
		container:    container,
	}
	wf, _ := container.GetMetadataStorage().GetWorkflowDefinition(wfName)
	flowMachine.flow = Convert(wf, flowId, container)
	flowCtx, err := container.GetClusterStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return nil, err
	}
	flowMachine.flowContext = flowCtx
	flowMachine.CurrentAction = flowMachine.flow.Actions[flowMachine.flowContext.CurrentAction]
	if flowMachine.flowContext.State == model.COMPLETED {
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
	f.CurrentAction = f.flow.Actions[wf.RootAction]
	f.FlowId = flowId
	dataMap := make(map[string]any)
	dataMap["input"] = input
	f.flowContext = &model.FlowContext{
		Id:            flowId,
		State:         model.RUNNING,
		CurrentAction: wf.RootAction,
		Data:          dataMap,
	}
	f.flowContext.State = model.RUNNING
	return f.saveContextAndDispatchAction(wf.RootAction, 1)
}

func (f *FlowMachine) MoveForwardAndDispatch(event string, dataMap map[string]any) (bool, error) {
	currentActionId := f.CurrentAction.GetId()
	nextActionMap := f.CurrentAction.GetNext()
	if f.completed {
		return true, nil
	}
	if nextActionMap == nil || len(nextActionMap) == 0 {
		f.MarkComplete()
		return true, nil
	}
	nextActionId := nextActionMap[event]
	f.CurrentAction = f.flow.Actions[nextActionId]
	f.flowContext.CurrentAction = nextActionId
	data := f.flowContext.Data
	if dataMap != nil || len(dataMap) > 0 {
		output := make(map[string]any)
		output["output"] = dataMap
		data[fmt.Sprintf("%d", currentActionId)] = output
	}
	f.flowContext.Data = data
	err := f.saveContextAndDispatchAction(f.CurrentAction.GetId(), 1)
	if err != nil {
		return false, err
	}
	return false, nil
}
func (f *FlowMachine) DispatchAction(actionId int, tryCount int) error {
	err := f.ValidateExecutionRequest(actionId)
	if err != nil {
		return err
	}
	act := &api.Action{
		WorkflowName: f.WorkflowName,
		FlowId:       f.FlowId,
		Data:         util.ConvertToProto(util.ResolveInputParams(f.flowContext, f.CurrentAction.GetInputParams())),
		ActionId:     int32(f.flowContext.CurrentAction),
		ActionName:   f.CurrentAction.GetName(),
		RetryCount:   int32(tryCount),
	}
	if f.CurrentAction.GetType() == action.ACTION_TYPE_SYSTEM {
		err = f.container.GetClusterStorage().DispatchAction(act, "system")
	} else {
		err = f.container.GetClusterStorage().DispatchAction(act, "user")
	}
	if err != nil {
		return err
	}
	return nil
}

func (f *FlowMachine) saveContextAndDispatchAction(actionId int, tryCount int) error {
	err := f.ValidateExecutionRequest(actionId)
	if err != nil {
		return err
	}
	act := &api.Action{
		WorkflowName: f.WorkflowName,
		FlowId:       f.FlowId,
		Data:         util.ConvertToProto(util.ResolveInputParams(f.flowContext, f.CurrentAction.GetInputParams())),
		ActionId:     int32(f.flowContext.CurrentAction),
		ActionName:   f.CurrentAction.GetName(),
		RetryCount:   int32(tryCount),
	}
	if f.CurrentAction.GetType() == action.ACTION_TYPE_SYSTEM {
		err = f.container.GetClusterStorage().SaveFlowContextAndDispatchAction(f.WorkflowName, f.FlowId, f.flowContext, act, "system")
	} else {
		err = f.container.GetClusterStorage().SaveFlowContextAndDispatchAction(f.WorkflowName, f.FlowId, f.flowContext, act, "user")
	}
	if err != nil {
		return err
	}
	return nil
}

func (f *FlowMachine) RetryAction(actionId int, tryCount int, retryAfter time.Duration) error {
	err := f.ValidateExecutionRequest(actionId)
	if err != nil {
		return err
	}
	act := &api.Action{
		WorkflowName: f.WorkflowName,
		FlowId:       f.FlowId,
		Data:         util.ConvertToProto(util.ResolveInputParams(f.flowContext, f.CurrentAction.GetInputParams())),
		ActionId:     int32(f.flowContext.CurrentAction),
		ActionName:   f.CurrentAction.GetName(),
		RetryCount:   int32(tryCount),
	}
	err = f.container.GetClusterStorage().Retry(act, retryAfter)

	if err != nil {
		return err
	}
	return nil
}

func (f *FlowMachine) DelayAction(actionId int, tryCount int, delay time.Duration) error {
	err := f.ValidateExecutionRequest(actionId)
	if err != nil {
		return err
	}
	act := &api.Action{
		WorkflowName: f.WorkflowName,
		FlowId:       f.FlowId,
		Data:         util.ConvertToProto(util.ResolveInputParams(f.flowContext, f.CurrentAction.GetInputParams())),
		ActionId:     int32(f.flowContext.CurrentAction),
		ActionName:   f.CurrentAction.GetName(),
		RetryCount:   int32(tryCount),
	}
	err = f.container.GetClusterStorage().Delay(act, delay)

	if err != nil {
		return err
	}
	return nil
}

func (f *FlowMachine) MoveForward(event string, dataMap map[string]any) (bool, error) {
	currentActionId := f.CurrentAction.GetId()
	nextActionMap := f.CurrentAction.GetNext()
	if f.completed {
		return true, nil
	}
	if nextActionMap == nil || len(nextActionMap) == 0 {
		f.MarkComplete()
		return true, nil
	}
	nextActionId := nextActionMap[event]
	f.CurrentAction = f.flow.Actions[nextActionId]
	f.flowContext.CurrentAction = nextActionId
	data := f.flowContext.Data
	if dataMap != nil || len(dataMap) > 0 {
		output := make(map[string]any)
		output["output"] = dataMap
		data[fmt.Sprintf("%d", currentActionId)] = output
	}
	f.flowContext.Data = data
	return false, f.container.GetClusterStorage().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
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
	return f.saveContextAndDispatchAction(f.CurrentAction.GetId(), 1)
}

func (f *FlowMachine) ExecuteSystemAction(tryCount int, actionId int) error {
	err := f.ValidateExecutionRequest(actionId)
	if err != nil {
		return err
	}
	currentAction := f.CurrentAction
	event, dataMap, err := currentAction.Execute(f.WorkflowName, f.flowContext, tryCount)
	if err != nil {
		return err
	}

	if currentAction.GetType() == action.ACTION_TYPE_SYSTEM {
		switch currentAction.GetName() {
		case "delay":
			f.MarkWaitingDelay()
			delay := f.CurrentAction.GetParams()["delay"].(time.Duration)
			f.DelayAction(f.CurrentAction.GetId(), 1, delay)
		case "wait":
			f.MarkWaitingEvent()
		default:
			completed, err := f.MoveForwardAndDispatch(event, dataMap)
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

func (f *FlowMachine) ValidateExecutionRequest(actionId int) error {
	if f.GetFlowState() == model.COMPLETED {
		return fmt.Errorf("can not run completed flow")
	}
	if f.GetFlowState() == model.FAILED {
		return fmt.Errorf("can not run failed flow")
	}
	if f.GetFlowState() == model.PAUSED {
		return fmt.Errorf("can not run paused flow")
	}
	if actionId != f.CurrentAction.GetId() {
		return fmt.Errorf("action %d already executed", actionId)
	}
	return nil
}
