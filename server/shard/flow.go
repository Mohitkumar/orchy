package shard

import (
	"fmt"
	"sync"

	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/metadata"
	"github.com/mohitkumar/orchy/server/model"
	"go.uber.org/zap"
)

type FlowStateMachineContainer struct {
	storage         Storage
	metadataService metadata.MetadataService
	stateMachines   map[string]*FlowStateMachine
	mu              sync.Mutex
}

func NewFlowStateMachineContainer(storage Storage, metadataService metadata.MetadataService) *FlowStateMachineContainer {
	return &FlowStateMachineContainer{
		stateMachines:   make(map[string]*FlowStateMachine),
		storage:         storage,
		metadataService: metadataService,
	}
}

func (sc *FlowStateMachineContainer) Store(wf string, flowId string, m *FlowStateMachine) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.stateMachines[fmt.Sprintf("%s:%s", wf, flowId)] = m
}

func (sc *FlowStateMachineContainer) Delete(wf string, flowId string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.stateMachines, fmt.Sprintf("%s:%s", wf, flowId))
}
func (sc *FlowStateMachineContainer) Init(wf string, flowId string, input map[string]any) (*FlowStateMachine, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	flow, err := sc.metadataService.GetFlow(wf, flowId)
	if err != nil {
		logger.Error("Workflow not found", zap.String("Workflow", wf), zap.Error(err))
		return nil, err
	}
	dataMap := make(map[string]any)
	dataMap["input"] = input
	flowCtx := &model.FlowContext{
		Id:               flowId,
		State:            model.RUNNING,
		Data:             dataMap,
		ExecutedActions:  map[int]bool{},
		CurrentActionIds: map[int]int{flow.RootAction: 1},
	}
	sm := NewFlowStateMachine(wf, flowId, flowCtx, flow)
	sc.stateMachines[fmt.Sprintf("%s:%s", wf, flowId)] = sm
	return sm, nil
}

func (sc *FlowStateMachineContainer) Get(wf string, flowId string) (*FlowStateMachine, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sm, ok := sc.stateMachines[fmt.Sprintf("%s:%s", wf, flowId)]
	flowCtx, err := sc.storage.GetFlowContext(wf, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not dispatch next action", zap.String("Workflow", wf), zap.String("FlowId", flowId), zap.Error(fmt.Errorf("workflow complted")))
		return nil, fmt.Errorf("flow already completed")
	}
	if !ok {
		flow, err := sc.metadataService.GetFlow(wf, flowId)
		if err != nil {
			logger.Error("Workflow not found", zap.String("Workflow", wf), zap.Error(err))
			return nil, err
		}
		sm = NewFlowStateMachine(wf, flowId, flowCtx, flow)
		sc.stateMachines[fmt.Sprintf("%s:%s", wf, flowId)] = sm
	} else {
		sm.context = flowCtx
	}
	return sm, nil
}

type FlowStateMachine struct {
	workflow string
	id       string
	context  *model.FlowContext
	flow     *flow.Flow
	mu       sync.Mutex
}

func NewFlowStateMachine(workflow string, id string, context *model.FlowContext, flow *flow.Flow) *FlowStateMachine {
	return &FlowStateMachine{
		workflow: workflow,
		id:       id,
		context:  context,
		flow:     flow,
	}
}

func (fm *FlowStateMachine) Validate(actionId int) error {
	if fm.context.State == model.COMPLETED || fm.context.State == model.FAILED {
		logger.Debug("flow already completed, can not dispatch next action", zap.String("Workflow", fm.workflow), zap.String("FlowId", fm.id), zap.Error(fmt.Errorf("workflow complted")))
		return fmt.Errorf("flow already completed")
	}
	if _, ok := fm.context.ExecutedActions[actionId]; ok {
		logger.Error("Action already executed", zap.String("Workflow", fm.workflow), zap.String("Id", fm.id), zap.Int("action", actionId))
		return fmt.Errorf("action %d already executed", actionId)
	}
	return nil
}

func (fm *FlowStateMachine) MoveForward(actionId int, event string, dataMap map[string]any) (bool, []int) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	currentAction := fm.flow.Actions[actionId]
	nextActionMap := currentAction.GetNext()
	if len(fm.context.ExecutedActions) == 0 {
		fm.context.ExecutedActions = make(map[int]bool)
	}
	fm.context.ExecutedActions[actionId] = true
	delete(fm.context.CurrentActionIds, actionId)
	if fm.isComplete() {
		fm.context.State = model.COMPLETED
		data := fm.context.Data
		if dataMap != nil || len(dataMap) > 0 {
			output := make(map[string]any)
			output["output"] = dataMap
			data[fmt.Sprintf("%d", actionId)] = output
		}
		fm.context.Data = data
		return true, []int{}
	}
	actionIdsToDispatch := nextActionMap[event]
	for _, actionId := range actionIdsToDispatch {
		fm.context.CurrentActionIds[actionId] = 1
	}
	data := fm.context.Data
	if dataMap != nil || len(dataMap) > 0 {
		output := make(map[string]any)
		output["output"] = dataMap
		data[fmt.Sprintf("%d", actionId)] = output
	}
	fm.context.Data = data
	return false, actionIdsToDispatch
}

func (fm *FlowStateMachine) ChangeState(state model.FlowState) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.context.State = state
}

func (fm *FlowStateMachine) isComplete() bool {
	allActions := make([]int, 0, len(fm.flow.Actions))
	for k := range fm.flow.Actions {
		allActions = append(allActions, k)
	}
	for _, actionId := range allActions {
		if _, ok := fm.context.ExecutedActions[actionId]; !ok {
			return false
		}
	}
	return true
}
