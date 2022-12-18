package flow

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mohitkumar/orchy/server/action"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
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
	wf, _ := container.GetWorkflowDao().Get(wfName)
	flowMachine.flow = Convert(wf, flowId, container)
	flowCtx, err := container.GetFlowDao().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(err))
		return nil, err
	}
	flowMachine.flowContext = flowCtx
	flowMachine.CurrentAction = flowMachine.flow.Actions[flowMachine.flowContext.CurrentAction]
	if flowMachine.flowContext.State == model.COMPLETED {
		flowMachine.completed = true
	}
	return flowMachine, nil
}

func (f *FlowMachine) Init(wfName string, input map[string]any) error {
	wf, err := f.container.GetWorkflowDao().Get(wfName)
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
	return f.MarkRunning()
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
	return false, f.container.GetFlowDao().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
}

func (f *FlowMachine) MarkComplete() {
	f.flowContext.State = model.COMPLETED
	f.completed = true
	f.container.GetFlowDao().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	successHandler := f.container.GetStateHandler().GetHandler(f.flow.SuccessHandler)
	err := successHandler(f.WorkflowName, f.FlowId)
	if err != nil {
		logger.Error("error in running success handler", zap.Error(err))
	}
	logger.Info("workflow completed", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
}

func (f *FlowMachine) MarkFailed() {
	f.flowContext.State = model.FAILED
	f.container.GetFlowDao().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	failureHandler := f.container.GetStateHandler().GetHandler(f.flow.FailureHandler)
	err := failureHandler(f.WorkflowName, f.FlowId)
	if err != nil {
		logger.Error("error in running failure handler", zap.Error(err))
	}
	logger.Info("workflow failed", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
}

func (f *FlowMachine) MarkWaitingDelay() {
	f.flowContext.State = model.WAITING_DELAY
	f.container.GetFlowDao().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	logger.Info("workflow waiting delay", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
}

func (f *FlowMachine) MarkWaitingEvent() {
	f.flowContext.State = model.WAITING_EVENT
	f.container.GetFlowDao().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	logger.Info("workflow waiting for event", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
}

func (f *FlowMachine) MarkPaused() error {
	f.flowContext.State = model.PAUSED
	err := f.container.GetFlowDao().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	if err != nil {
		return err
	}
	logger.Info("workflow paused", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
	return nil
}

func (f *FlowMachine) MarkRunning() error {
	f.flowContext.State = model.RUNNING
	err := f.container.GetFlowDao().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
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
	f.MarkRunning()
	return f.Execute(1, f.CurrentAction.GetId())
}

func (f *FlowMachine) Execute(tryCount int, actionId int) error {
	err := f.ValidateExecutionRequest(actionId)
	if err != nil {
		return err
	}
	currentAction := f.CurrentAction
	if currentAction.GetType() == action.ACTION_TYPE_SYSTEM {
		return f.pushSystemAction(tryCount)
	} else {
		_, _, err := currentAction.Execute(f.WorkflowName, f.flowContext, tryCount)
		if err != nil {
			return err
		}
	}

	return nil
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
		case "wait":
			f.MarkWaitingEvent()
		default:
			completed, err := f.MoveForward(event, dataMap)
			if completed {
				return nil
			}
			if err != nil {
				logger.Error("error moving forward in workflow", zap.Error(err))
				return err
			}
			return f.Execute(1, f.CurrentAction.GetId())
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

func (f *FlowMachine) pushSystemAction(tryCount int) error {
	req := model.ActionExecutionRequest{
		WorkflowName: f.WorkflowName,
		ActionId:     f.CurrentAction.GetId(),
		FlowId:       f.FlowId,
		TryNumber:    tryCount,
		TaskName:     f.CurrentAction.GetName(),
	}
	data, _ := f.container.ActionExecutionRequestEncDec.Encode(req)

	err := f.container.GetQueue().Push("system", f.FlowId, data)
	if err != nil {
		return err
	}
	return nil
}
