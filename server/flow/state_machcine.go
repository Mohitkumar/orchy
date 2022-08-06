package flow

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mohitkumar/orchy/server/action"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/util"
)

type FlowMachine struct {
	WorkflowName  string
	FlowId        string
	flow          *Flow
	flowContext   *model.FlowContext
	container     *container.DIContiner
	CurrentAction action.Action
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
	wf, _ := f.container.GetWorkflowDao().Get(wfName)
	flowId := uuid.New().String()

	f.flow = Convert(wf, flowId, f.container)
	f.CurrentAction = f.flow.Actions[wf.RootAction]

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

func (f *FlowMachine) MoveForward(event string, dataMap map[string]any) error {
	currentActionId := f.CurrentAction.GetId()
	nextActionMap := f.CurrentAction.GetNext()
	if nextActionMap == nil || len(nextActionMap) == 0 {
		f.flowContext.State = model.COMPLETED
		f.container.GetFlowDao().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
		return fmt.Errorf("no more action to execute")
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
	return f.container.GetFlowDao().SaveFlowContext(f.WorkflowName, f.FlowId, f.flowContext)
}

func (f *FlowMachine) Execute(retryCount int) error {
	currentAction := f.CurrentAction
	event, dataMap, err := currentAction.Execute(f.WorkflowName, f.flowContext, retryCount)
	if err != nil {
		return err
	}
	if currentAction.GetType() == action.ACTION_TYPE_SYSTEM {
		switch currentAction.GetName() {
		case "switch":
			f.MoveForward(event, dataMap)
			return f.Execute(1)
		case "delay":

		}
	}
	return nil
}