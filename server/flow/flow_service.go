package flow

import (
	"fmt"
	"sync"
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

type FlowService struct {
	container          *container.DIContiner
	executionChannel   chan model.FlowExecutionRequest
	stateChangeChannel chan model.FlowStateChangeRequest
	stop               chan struct{}
	wg                 *sync.WaitGroup
}

func NewFlowService(container *container.DIContiner, wg *sync.WaitGroup) *FlowService {
	return &FlowService{
		container:          container,
		executionChannel:   make(chan model.FlowExecutionRequest, 10000),
		stateChangeChannel: make(chan model.FlowStateChangeRequest, 1000),
		stop:               make(chan struct{}),
		wg:                 wg,
	}
}

func (f *FlowService) ExecuteAction(wfName string, wfId string, event string, actionId int, tryCount int, data map[string]any) {
	req := model.FlowExecutionRequest{
		WorkflowName: wfName,
		FlowId:       wfId,
		ActionId:     actionId,
		Event:        event,
		DataMap:      data,
	}
	f.executionChannel <- req
}

func (f *FlowService) Stop() error {
	f.stop <- struct{}{}
	return nil
}

func (f *FlowService) Start() {
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		for {
			select {
			case req := <-f.executionChannel:
				f.execute(req)
			case req := <-f.stateChangeChannel:
				f.changeState(req.WorkflowName, req.FlowId, req.State)
			case <-f.stop:
				logger.Info("stopping flow execution service")
				return
			}
		}
	}()
}

func (f *FlowService) Init(wfName string, input map[string]any) (string, error) {
	wf, err := f.container.GetMetadataStorage().GetWorkflowDefinition(wfName)
	if err != nil {
		logger.Error("Workflow Definition not found", zap.Error(err))
		return "", err
	}
	flowId := uuid.New().String()
	flow := Convert(wf, flowId, f.container)
	dataMap := make(map[string]any)
	dataMap["input"] = input
	flowCtx := &model.FlowContext{
		Id:              flowId,
		State:           model.RUNNING,
		Data:            dataMap,
		ExecutedActions: map[int]bool{wf.RootAction: true},
	}
	f.saveContextAndDispatchAction(wfName, flowId, []int{wf.RootAction}, 1, flow, flowCtx)
	if err != nil {
		logger.Error("error executiong flow", zap.String("workflowName", wfName), zap.String("flowId", flowId), zap.Error(err))
		return flowId, err
	}
	return flowId, nil
}

func (f *FlowService) execute(req model.FlowExecutionRequest) {
	wf, err := f.container.GetMetadataStorage().GetWorkflowDefinition(req.WorkflowName)
	if err != nil {
		logger.Error("Workflow Definition not found", zap.Error(err))
		return
	}
	flow := Convert(wf, req.FlowId, f.container)
	flowCtx, err := f.container.GetClusterStorage().GetFlowContext(req.WorkflowName, req.FlowId)
	if err != nil {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	if flowCtx.State == model.COMPLETED || flowCtx.State == model.FAILED {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	if _, ok := flowCtx.ExecutedActions[req.ActionId]; ok {
		logger.Error("Action already executed", zap.Any("executionRequest", req))
		return
	}
	currentAction := flow.Actions[req.ActionId]
	nextActionMap := currentAction.GetNext()
	flowCtx.ExecutedActions[req.ActionId] = true
	if f.isComplete(flow, flowCtx) {
		f.markComplete(req.WorkflowName, req.FlowId, flow, flowCtx)
	}
	actionIdsToDispatch := nextActionMap[req.Event]
	data := flowCtx.Data
	if req.DataMap != nil || len(req.DataMap) > 0 {
		output := make(map[string]any)
		output["output"] = req.DataMap
		data[fmt.Sprintf("%d", req.ActionId)] = output
	}
	flowCtx.Data = data
	err = f.saveContextAndDispatchAction(req.WorkflowName, req.FlowId, actionIdsToDispatch, 1, flow, flowCtx)
	if err != nil {
		logger.Error("error executiong flow", zap.Any("flowRequest", req), zap.Error(err))
	}

}

func (f *FlowService) saveContextAndDispatchAction(wfName string, flowId string, actionIds []int, tryCount int, flow *Flow, flowCtx *model.FlowContext) error {
	var actions []*api.Action
	for _, actionId := range actionIds {
		currentAction := flow.Actions[actionId]
		var actionType api.Action_Type
		if currentAction.GetType() == action.ACTION_TYPE_SYSTEM {
			actionType = api.Action_SYSTEM
		} else {
			actionType = api.Action_USER
		}
		act := &api.Action{
			WorkflowName: wfName,
			FlowId:       flowId,
			Data:         util.ConvertToProto(util.ResolveInputParams(flowCtx, currentAction.GetInputParams())),
			ActionId:     int32(actionId),
			ActionName:   currentAction.GetName(),
			RetryCount:   int32(tryCount),
			Type:         actionType,
		}
		actions = append(actions, act)
	}
	err := f.container.GetClusterStorage().SaveFlowContextAndDispatchAction(wfName, flowId, flowCtx, actions)
	if err != nil {
		return err
	}
	return nil
}

func (f *FlowService) isComplete(flow *Flow, flowCtx *model.FlowContext) bool {
	allActions := make([]int, 0, len(flow.Actions))
	for k := range flow.Actions {
		allActions = append(allActions, k)
	}
	for _, actionId := range allActions {
		if _, ok := flowCtx.ExecutedActions[actionId]; !ok {
			return false
		}
	}
	return true
}

func (f *FlowService) markComplete(wfName string, flowId string, flow *Flow, flowCtx *model.FlowContext) {
	flowCtx.State = model.COMPLETED
	f.container.GetClusterStorage().SaveFlowContext(wfName, flowId, flowCtx)
	successHandler := f.container.GetStateHandler().GetHandler(flow.SuccessHandler)
	err := successHandler(wfName, flowId)
	if err != nil {
		logger.Error("error in running success handler", zap.Error(err))
	}
	logger.Info("workflow completed", zap.String("workflow", wfName), zap.String("id", flowId))
}

func (f *FlowService) ExecuteSystemAction(wfName string, flowId string, tryCount int, actionId int) {
	wf, err := f.container.GetMetadataStorage().GetWorkflowDefinition(wfName)
	if err != nil {
		logger.Error("Workflow Definition not found", zap.Error(err))
		return
	}
	flow := Convert(wf, flowId, f.container)
	flowCtx, err := f.container.GetClusterStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	if flowCtx.State == model.COMPLETED || flowCtx.State == model.FAILED {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	if _, ok := flowCtx.ExecutedActions[actionId]; ok {
		logger.Error("Action already executed", zap.String("Workflow", wfName), zap.String("Id", flowId), zap.Int("action", actionId))
		return
	}
	currentAction := flow.Actions[actionId]
	event, dataMap, err := currentAction.Execute(wfName, flowCtx, tryCount)
	if err != nil {
		logger.Error("error executing workflow", zap.String("Workflow", wfName), zap.String("Id", flowId), zap.Int("action", actionId))
	}

	if currentAction.GetType() == action.ACTION_TYPE_SYSTEM {
		switch currentAction.GetName() {
		case "delay":
			f.MarkWaitingDelay(wfName, flowId)
			delay := currentAction.GetParams()["delay"].(time.Duration)
			f.DelayAction(wfName, flowId, currentAction.GetId(), 1, delay)
		case "wait":
			f.MarkWaitingEvent(wfName, flowId)
		default:
			f.ExecuteAction(wfName, flowId, event, actionId, 1, dataMap)
		}
	} else {
		logger.Error("should be system action")
	}
}

func (f *FlowService) RetryAction(wfName string, flowId string, actionId int, tryCount int, retryAfter time.Duration) {
	wf, err := f.container.GetMetadataStorage().GetWorkflowDefinition(wfName)
	if err != nil {
		logger.Error("Workflow Definition not found", zap.Error(err))
		return
	}
	flow := Convert(wf, flowId, f.container)
	flowCtx, err := f.container.GetClusterStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	if flowCtx.State == model.COMPLETED || flowCtx.State == model.FAILED {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	if _, ok := flowCtx.ExecutedActions[actionId]; ok {
		logger.Error("Action already executed", zap.String("Workflow", wfName), zap.String("Id", flowId), zap.Int("action", actionId))
		return
	}
	currentAction := flow.Actions[actionId]
	act := &api.Action{
		WorkflowName: wfName,
		FlowId:       flowId,
		Data:         util.ConvertToProto(util.ResolveInputParams(flowCtx, currentAction.GetInputParams())),
		ActionId:     int32(currentAction.GetId()),
		ActionName:   currentAction.GetName(),
		RetryCount:   int32(tryCount),
	}
	err = f.container.GetClusterStorage().Retry(act, retryAfter)
	if err != nil {
		logger.Error("error retrying workflow", zap.String("Workflow", wfName), zap.String("Id", flowId), zap.Int("action", actionId))
	}
}

func (f *FlowService) DelayAction(wfName string, flowId string, actionId int, tryCount int, delay time.Duration) {
	wf, err := f.container.GetMetadataStorage().GetWorkflowDefinition(wfName)
	if err != nil {
		logger.Error("Workflow Definition not found", zap.Error(err))
		return
	}
	flow := Convert(wf, flowId, f.container)
	flowCtx, err := f.container.GetClusterStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	if flowCtx.State == model.COMPLETED || flowCtx.State == model.FAILED {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	if _, ok := flowCtx.ExecutedActions[actionId]; ok {
		logger.Error("Action already executed", zap.String("Workflow", wfName), zap.String("Id", flowId), zap.Int("action", actionId))
		return
	}
	currentAction := flow.Actions[actionId]
	act := &api.Action{
		WorkflowName: wfName,
		FlowId:       flowId,
		Data:         util.ConvertToProto(util.ResolveInputParams(flowCtx, currentAction.GetInputParams())),
		ActionId:     int32(currentAction.GetId()),
		ActionName:   currentAction.GetName(),
		RetryCount:   int32(tryCount),
	}
	err = f.container.GetClusterStorage().Delay(act, delay)

	if err != nil {
		logger.Error("error running delay action", zap.String("Workflow", wfName), zap.String("Id", flowId), zap.Int("action", actionId))
	}
}

func (f *FlowService) DispatchAction(wfName string, flowId string, actionId int, tryCount int) {
	wf, err := f.container.GetMetadataStorage().GetWorkflowDefinition(wfName)
	if err != nil {
		logger.Error("Workflow Definition not found", zap.Error(err))
		return
	}
	flow := Convert(wf, flowId, f.container)
	flowCtx, err := f.container.GetClusterStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	if flowCtx.State == model.COMPLETED || flowCtx.State == model.FAILED {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	if _, ok := flowCtx.ExecutedActions[actionId]; ok {
		logger.Error("Action already executed", zap.String("Workflow", wfName), zap.String("Id", flowId), zap.Int("action", actionId))
		return
	}
	currentAction := flow.Actions[actionId]
	var actionType api.Action_Type
	if currentAction.GetType() == action.ACTION_TYPE_SYSTEM {
		actionType = api.Action_SYSTEM
	} else {
		actionType = api.Action_USER
	}
	act := &api.Action{
		WorkflowName: wfName,
		FlowId:       flowId,
		Data:         util.ConvertToProto(util.ResolveInputParams(flowCtx, currentAction.GetInputParams())),
		ActionId:     int32(currentAction.GetId()),
		ActionName:   currentAction.GetName(),
		RetryCount:   int32(tryCount),
		Type:         actionType,
	}

	err = f.container.GetClusterStorage().DispatchAction(flowId, []*api.Action{act})
	if err != nil {
		logger.Error("error dispatching action", zap.String("Workflow", wfName), zap.String("Id", flowId), zap.Int("action", actionId))
	}
}

func (f *FlowService) MarkFailed(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.FAILED,
	}
}

func (f *FlowService) MarkWaitingDelay(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.FAILED,
	}
	logger.Info("workflow waiting delay", zap.String("workflow", wfName), zap.String("id", flowId))
}

func (f *FlowService) MarkWaitingEvent(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.FAILED,
	}
	logger.Info("workflow waiting for event", zap.String("workflow", wfName), zap.String("id", flowId))
}

func (f *FlowService) MarkRunning(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.FAILED,
	}
	logger.Info("workflow running", zap.String("workflow", wfName), zap.String("id", flowId))
}

func (f *FlowService) changeState(wfName string, flowId string, state model.FlowState) {
	flowCtx, err := f.container.GetClusterStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	flowCtx.State = state
	f.container.GetClusterStorage().SaveFlowContext(wfName, flowId, flowCtx)
	if state == model.FAILED {
		wf, err := f.container.GetMetadataStorage().GetWorkflowDefinition(wfName)
		if err != nil {
			logger.Error("Workflow Definition not found", zap.Error(err))
			return
		}
		flow := Convert(wf, flowId, f.container)
		failureHandler := f.container.GetStateHandler().GetHandler(flow.FailureHandler)
		err = failureHandler(wfName, flowId)
		if err != nil {
			logger.Error("error in running failure handler", zap.Error(err))
		}
		logger.Info("workflow failed", zap.String("workflow", wfName), zap.String("id", flowId))
	} else if state == model.COMPLETED {
		wf, err := f.container.GetMetadataStorage().GetWorkflowDefinition(wfName)
		if err != nil {
			logger.Error("Workflow Definition not found", zap.Error(err))
			return
		}
		flow := Convert(wf, flowId, f.container)
		successHandler := f.container.GetStateHandler().GetHandler(flow.SuccessHandler)
		err = successHandler(wfName, flowId)
		if err != nil {
			logger.Error("error in running success handler", zap.Error(err))
		}
		logger.Info("workflow completed", zap.String("workflow", wfName), zap.String("id", flowId))
	}
}
