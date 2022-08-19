package flow

import (
	"fmt"

	"github.com/google/uuid"
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

func GetFlowStateMachine(wfName string, flowId string, container *container.DIContiner) *FlowMachine {
	flowMachine := &FlowMachine{
		WorkflowName: wfName,
		FlowId:       flowId,
		container:    container,
	}
	wf, _ := container.GetWorkflowDao().Get(wfName)
	flowMachine.flow = Convert(wf, flowId, container)
	flowMachine.flowContext, _ = container.GetFlowDao().GetFlowContext(wfName, flowId)
	flowMachine.CurrentAction = flowMachine.flow.Actions[flowMachine.flowContext.CurrentAction]
	return flowMachine
}

func (f *FlowMachine) Init(wfName string, input map[string]any) error {
	wf, err := f.container.GetWorkflowDao().Get(wfName)
	if err != nil {
		return fmt.Errorf("workflow %s not found", wfName)
	}
	flowId := uuid.New().String()

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
	return f.container.GetFlowDao().SaveFlowContext(wfName, flowId, f.flowContext)
}

func (f *FlowMachine) MoveForward(event string, dataMap map[string]any) (bool, error) {
	currentActionId := f.CurrentAction.GetId()
	nextActionMap := f.CurrentAction.GetNext()
	if nextActionMap == nil || len(nextActionMap) == 0 {
		f.MarkComplete()
		return true, nil
	}
	nextActionId := nextActionMap[event]
	f.CurrentAction = f.flow.Actions[nextActionId]
	f.flowContext.CurrentAction = nextActionId
	if dataMap != nil || len(dataMap) > 0 {
		data := f.flowContext.Data
		output := make(map[string]any)
		output["output"] = dataMap
		data[fmt.Sprintf("%d", currentActionId)] = util.ConvertMapToStructPb(output)
	}
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
	f.flowContext.State = model.COMPLETED
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
	logger.Info("workflow wairing delay", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
}

func (f *FlowMachine) MarkWaitingEvent() {
	f.flowContext.State = model.WAITING_EVENT
	f.container.GetFlowDao().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	logger.Info("workflow wairing for event", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
}

func (f *FlowMachine) MarkPaused() {
	f.flowContext.State = model.PAUSED
	f.container.GetFlowDao().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	logger.Info("workflow paused", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
}

func (f *FlowMachine) MarkRunning() {
	f.flowContext.State = model.RUNNING
	f.container.GetFlowDao().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
	logger.Info("workflow running", zap.String("workflow", f.WorkflowName), zap.String("id", f.FlowId))
}

func (f *FlowMachine) Execute(tryCount int) error {
	currentAction := f.CurrentAction
	event, dataMap, err := currentAction.Execute(f.WorkflowName, f.flowContext, tryCount)
	if err != nil {
		return err
	}
	if currentAction.GetType() == action.ACTION_TYPE_SYSTEM {
		switch currentAction.GetName() {
		case "switch":
			completed, err := f.MoveForward(event, dataMap)
			if err != nil {
				logger.Error("error moving forward in workflow", zap.Error(err))
				return err
			}
			if completed {
				logger.Info("workflow completed, no more action to execute", zap.String("workflow", f.WorkflowName), zap.String("flow", f.FlowId))
				return nil
			}
			return f.Execute(1)
		case "delay":
			f.MarkWaitingDelay()
		}
	}
	return nil
}

func (f *FlowMachine) DelayTask() {
	completed, err := f.MoveForward("default", nil)
	if err != nil {
		logger.Error("error moving forward in workflow", zap.Error(err))
		return
	}
	if completed {
		logger.Info("workflow completed, no more action to execute", zap.String("workflow", f.WorkflowName), zap.String("flow", f.FlowId))
		return
	}
	f.Execute(1)
}

func (f *FlowMachine) RetryTask(tryNumber int) {
	if f.completed {
		logger.Info("workflow completed, can not retry", zap.String("workflow", f.WorkflowName), zap.String("flow", f.FlowId))
		return
	}
	f.Execute(tryNumber)
}
