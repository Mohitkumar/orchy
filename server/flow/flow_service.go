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
		TryCount:     tryCount,
		RequestType:  model.NEW_FLOW_EXECUTION,
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
				switch req.RequestType {
				case model.RETRY_FLOW_EXECUTION:
					f.retry(req.WorkflowName, req.FlowId, req.ActionId, req.TryCount)
				case model.RESUME_FLOW_EXECUTION:
					f.resume(req.WorkflowName, req.FlowId, req.TryCount)
				default:
					f.execute(req)
				}
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
		logger.Error("Workflow Definition not found", zap.String("Workflow", wfName), zap.Error(err))
		return "", err
	}
	flowId := uuid.New().String()
	flow := Convert(wf, flowId, f.container)
	dataMap := make(map[string]any)
	dataMap["input"] = input
	flowCtx := &model.FlowContext{
		Id:               flowId,
		State:            model.RUNNING,
		Data:             dataMap,
		ExecutedActions:  map[int]bool{wf.RootAction: true},
		CurrentActionIds: map[int]int{wf.RootAction: 1},
	}
	f.saveContextAndDispatchAction(wfName, flowId, []int{wf.RootAction}, 1, flow, flowCtx)
	if err != nil {
		logger.Error("error executiong flow", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(err))
		return flowId, err
	}
	return flowId, nil
}

func (f *FlowService) execute(req model.FlowExecutionRequest) {
	flow, flowCtx, err := f.validateAndGetFlow(req.WorkflowName, req.FlowId)
	if err != nil {
		return
	}
	currentAction := flow.Actions[req.ActionId]
	nextActionMap := currentAction.GetNext()
	flowCtx.ExecutedActions[req.ActionId] = true
	delete(flowCtx.CurrentActionIds, req.ActionId)
	if f.isComplete(flow, flowCtx) {
		f.markComplete(req.WorkflowName, req.FlowId, flow, flowCtx)
	}
	actionIdsToDispatch := nextActionMap[req.Event]
	for _, actionId := range actionIdsToDispatch {
		flowCtx.CurrentActionIds[actionId] = 1
	}
	data := flowCtx.Data
	if req.DataMap != nil || len(req.DataMap) > 0 {
		output := make(map[string]any)
		output["output"] = req.DataMap
		data[fmt.Sprintf("%d", req.ActionId)] = output
	}
	flowCtx.Data = data
	err = f.saveContextAndDispatchAction(req.WorkflowName, req.FlowId, actionIdsToDispatch, 1, flow, flowCtx)
	if err != nil {
		logger.Error("error executiong flow", zap.Any("Flow", req), zap.Error(err))
	}

}

func (f *FlowService) validateAndGetFlow(wfName string, flowId string) (*Flow, *model.FlowContext, error) {
	wf, err := f.container.GetMetadataStorage().GetWorkflowDefinition(wfName)
	if err != nil {
		logger.Error("Workflow Definition not found", zap.String("Workflow", wfName), zap.Error(err))
		return nil, nil, err
	}
	flow := Convert(wf, flowId, f.container)
	flowCtx, err := f.container.GetClusterStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not create flow machine", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(fmt.Errorf("workflow complted")))
		return nil, nil, fmt.Errorf("flow already completed")
	}
	if flowCtx.State == model.COMPLETED || flowCtx.State == model.FAILED {
		logger.Debug("flow already completed, can not create flow machine", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(fmt.Errorf("workflow complted")))
		return nil, nil, fmt.Errorf("flow already completed")
	}
	return flow, flowCtx, nil
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
	logger.Info("workflow completed", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowService) ExecuteSystemAction(wfName string, flowId string, tryCount int, actionId int) {
	flow, flowCtx, err := f.validateAndGetFlow(wfName, flowId)
	if err != nil {
		return
	}
	currentAction := flow.Actions[actionId]
	event, dataMap, err := currentAction.Execute(wfName, flowCtx, tryCount)
	if err != nil {
		logger.Error("error executing workflow", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
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
	flow, flowCtx, err := f.validateAndGetFlow(wfName, flowId)
	if err != nil {
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
		logger.Error("error retrying workflow", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
	}
}

func (f *FlowService) DelayAction(wfName string, flowId string, actionId int, tryCount int, delay time.Duration) {
	flow, flowCtx, err := f.validateAndGetFlow(wfName, flowId)
	if err != nil {
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
		logger.Error("error running delay action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
	}
}

func (f *FlowService) IsValid(wfName string, flowId string, actionId int) bool {
	flowCtx, err := f.container.GetClusterStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return false
	}
	if flowCtx.State == model.COMPLETED || flowCtx.State == model.FAILED {
		logger.Debug("flow already completed, can not create flow machine", zap.Error(fmt.Errorf("workflow complted")))
		return false
	}
	if _, ok := flowCtx.ExecutedActions[actionId]; ok {
		logger.Error("Action already executed", zap.String("Workflow", wfName), zap.String("Id", flowId), zap.Int("action", actionId))
		return false
	}
	return true
}

func (f *FlowService) ExecuteRetry(wfName string, flowId string, actionId int, tryCount int) {
	req := model.FlowExecutionRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		ActionId:     actionId,
		TryCount:     tryCount,
		RequestType:  model.RETRY_FLOW_EXECUTION,
	}
	f.executionChannel <- req
}

func (f *FlowService) ExecuteResume(wfName string, flowId string, event string) {
	req := model.FlowExecutionRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		Event:        event,
		TryCount:     1,
		RequestType:  model.RESUME_FLOW_EXECUTION,
	}
	f.executionChannel <- req
}

func (f *FlowService) retry(wfName string, flowId string, actionId int, tryCount int) {
	flow, flowCtx, err := f.validateAndGetFlow(wfName, flowId)
	if err != nil {
		return
	}

	flowCtx.CurrentActionIds[actionId] = tryCount
	err = f.saveContextAndDispatchAction(wfName, flowId, []int{actionId}, tryCount, flow, flowCtx)
	if err != nil {
		logger.Error("error dispatching action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
	}
}

func (f *FlowService) resume(wfName string, flowId string, tryCount int) {
	flow, flowCtx, err := f.validateAndGetFlow(wfName, flowId)
	if err != nil {
		return
	}
	actionIds := make([]int, 0, len(flowCtx.CurrentActionIds))
	for k := range flowCtx.CurrentActionIds {
		actionIds = append(actionIds, k)
	}

	err = f.saveContextAndDispatchAction(wfName, flowId, actionIds, tryCount, flow, flowCtx)
	if err != nil {
		logger.Error("error dispatching action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Any("actions", actionIds))
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
	logger.Info("workflow waiting delay", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowService) MarkWaitingEvent(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.FAILED,
	}
	logger.Info("workflow waiting for event", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowService) MarkRunning(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.FAILED,
	}
	logger.Info("workflow running", zap.String("workflow", wfName), zap.String("FlowId", flowId))
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
		logger.Info("workflow failed", zap.String("workflow", wfName), zap.String("FlowId", flowId))
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
		logger.Info("workflow completed", zap.String("workflow", wfName), zap.String("FlowId", flowId))
	}
}
