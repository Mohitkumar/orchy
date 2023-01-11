package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mohitkumar/orchy/server/action"
	"github.com/mohitkumar/orchy/server/cluster"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/metadata"
	"github.com/mohitkumar/orchy/server/model"
	"go.uber.org/zap"
)

type FlowEngine struct {
	cluster            *cluster.Cluster
	metadataService    metadata.MetadataService
	executionChannel   chan model.FlowExecutionRequest
	stateChangeChannel chan model.FlowStateChangeRequest
	stop               chan struct{}
	wg                 *sync.WaitGroup
}

func NewFlowEngine(cluster *cluster.Cluster, metadadataService metadata.MetadataService, wg *sync.WaitGroup) *FlowEngine {
	return &FlowEngine{
		cluster:            cluster,
		metadataService:    metadadataService,
		executionChannel:   make(chan model.FlowExecutionRequest, 1000),
		stateChangeChannel: make(chan model.FlowStateChangeRequest, 1000),
		stop:               make(chan struct{}),
		wg:                 wg,
	}
}

func (f *FlowEngine) GetExecutionChannel() chan model.FlowExecutionRequest {
	return f.executionChannel
}

func (f *FlowEngine) ExecuteAction(wfName string, wfId string, event string, actionId int, data map[string]any) {
	req := model.FlowExecutionRequest{
		WorkflowName: wfName,
		FlowId:       wfId,
		ActionId:     actionId,
		Event:        event,
		DataMap:      data,
		RequestType:  model.NEW_FLOW_EXECUTION,
	}
	f.executionChannel <- req
}

func (f *FlowEngine) Stop() error {
	f.stop <- struct{}{}
	return nil
}

func (f *FlowEngine) Start() {
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		for {
			select {
			case req := <-f.executionChannel:
				switch req.RequestType {
				case model.RETRY_FLOW_EXECUTION:
					f.retry(req.WorkflowName, req.FlowId, req.ActionId)
				case model.RESUME_FLOW_EXECUTION:
					f.resume(req.WorkflowName, req.FlowId)
				case model.SYSTEM_FLOW_EXECUTION:
					f.executeSystemAction(req.WorkflowName, req.FlowId, req.ActionId)
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

func (f *FlowEngine) Init(wfName string, input map[string]any) (string, error) {
	flowId := uuid.New().String()
	flow, err := f.metadataService.GetFlow(wfName, flowId)
	if err != nil {
		logger.Error("Workflow not found", zap.String("Workflow", wfName), zap.Error(err))
		return "", err
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
	f.saveContextAndDispatchAction(wfName, flowId, []int{flow.RootAction}, flow, flowCtx)
	if err != nil {
		logger.Error("error executiong flow", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(err))
		return flowId, err
	}
	return flowId, nil
}

func (f *FlowEngine) execute(req model.FlowExecutionRequest) {
	flow, flowCtx, err := f.validateAndGetFlow(req.WorkflowName, req.FlowId, req.ActionId)
	if err != nil {
		return
	}
	currentAction := flow.Actions[req.ActionId]
	nextActionMap := currentAction.GetNext()
	if len(flowCtx.ExecutedActions) == 0 {
		flowCtx.ExecutedActions = make(map[int]bool)
	}
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
	err = f.saveContextAndDispatchAction(req.WorkflowName, req.FlowId, actionIdsToDispatch, flow, flowCtx)
	if err != nil {
		logger.Error("error executiong flow", zap.Any("Flow", req), zap.Error(err))
	}

}

func (f *FlowEngine) validateFlow(wfName string, flowId string, actionId int) error {
	flowCtx, err := f.cluster.GetStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not dispatch next action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(fmt.Errorf("workflow complted")))
		return fmt.Errorf("flow already completed")
	}
	if flowCtx.State == model.COMPLETED || flowCtx.State == model.FAILED {
		logger.Debug("flow already completed, can not dispatch next action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(fmt.Errorf("workflow complted")))
		return fmt.Errorf("flow already completed")
	}
	if _, ok := flowCtx.ExecutedActions[actionId]; ok {
		logger.Error("Action already executed", zap.String("Workflow", wfName), zap.String("Id", flowId), zap.Int("action", actionId))
		return fmt.Errorf("action %d already executed", actionId)
	}
	return nil
}

func (f *FlowEngine) validateAndGetFlow(wfName string, flowId string, actionId int) (*flow.Flow, *model.FlowContext, error) {
	flow, err := f.metadataService.GetFlow(wfName, flowId)
	if err != nil {
		logger.Error("Workflow not found", zap.String("Workflow", wfName), zap.Error(err))
		return nil, nil, err
	}
	flowCtx, err := f.cluster.GetStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not dispatch next action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(fmt.Errorf("workflow complted")))
		return nil, nil, fmt.Errorf("flow already completed")
	}
	if flowCtx.State == model.COMPLETED || flowCtx.State == model.FAILED {
		logger.Debug("flow already completed, can not dispatch next action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(fmt.Errorf("workflow complted")))
		return nil, nil, fmt.Errorf("flow already completed")
	}
	if _, ok := flowCtx.ExecutedActions[actionId]; ok {
		logger.Error("Action already executed", zap.String("Workflow", wfName), zap.String("Id", flowId), zap.Int("action", actionId))
		return nil, nil, fmt.Errorf("action %d already executed", actionId)
	}
	return flow, flowCtx, nil
}
func (f *FlowEngine) saveContextAndDispatchAction(wfName string, flowId string, actionIds []int, flow *flow.Flow, flowCtx *model.FlowContext) error {
	var actions []model.ActionExecutionRequest
	for _, actionId := range actionIds {
		currentAction := flow.Actions[actionId]
		var actionType model.ActionType
		if currentAction.GetType() == action.ACTION_TYPE_SYSTEM {
			actionType = model.ACTION_TYPE_SYSTEM
		} else {
			actionType = model.ACTION_TYPE_USER
		}
		act := model.ActionExecutionRequest{
			ActionId:   actionId,
			ActionType: actionType,
			ActionName: currentAction.GetName(),
		}
		actions = append(actions, act)
	}
	err := f.cluster.GetStorage().SaveFlowContextAndDispatchAction(wfName, flowId, flowCtx, actions)
	if err != nil {
		return err
	}
	return nil
}

func (f *FlowEngine) isComplete(flow *flow.Flow, flowCtx *model.FlowContext) bool {
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

func (f *FlowEngine) markComplete(wfName string, flowId string, flow *flow.Flow, flowCtx *model.FlowContext) {
	flowCtx.State = model.COMPLETED
	f.cluster.GetStorage().SaveFlowContext(wfName, flowId, flowCtx)
	successHandler := f.cluster.GetStateHandler().GetHandler(flow.SuccessHandler)
	err := successHandler(wfName, flowId)
	if err != nil {
		logger.Error("error in running success handler", zap.Error(err))
	}
	logger.Info("workflow completed", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowEngine) executeSystemAction(wfName string, flowId string, actionId int) {
	flow, flowCtx, err := f.validateAndGetFlow(wfName, flowId, actionId)
	if err != nil {
		return
	}
	currentAction := flow.Actions[actionId]
	event, dataMap, err := currentAction.Execute(wfName, flowCtx, 1)
	if err != nil {
		logger.Error("error executing workflow", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId), zap.Error(err))
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
			f.ExecuteAction(wfName, flowId, event, actionId, dataMap)
		}
	} else {
		logger.Error("should be system action")
	}
}

func (f *FlowEngine) RetryAction(wfName string, flowId string, actionId int, reason string) {
	flow, flowCtx, err := f.validateAndGetFlow(wfName, flowId, actionId)
	if err != nil {
		return
	}
	action := flow.Actions[actionId]
	actionDefinition, err := f.metadataService.GetMetadataStorage().GetActionDefinition(action.GetName())
	if err != nil {
		logger.Error("action definition not found ", zap.String("action", action.GetName()), zap.Error(err))
		return
	}
	tryCount := flowCtx.CurrentActionIds[actionId]
	logger.Info("retrying action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.String("reason", reason), zap.Int("action", actionId), zap.Int("retry", tryCount+1))
	if tryCount < actionDefinition.RetryCount {
		var retryAfter time.Duration
		switch actionDefinition.RetryPolicy {
		case model.RETRY_POLICY_FIXED:
			retryAfter = time.Duration(actionDefinition.RetryAfterSeconds) * time.Second
		case model.RETRY_POLICY_BACKOFF:
			retryAfter = time.Duration(actionDefinition.RetryAfterSeconds*(tryCount+1)) * time.Second
		}
		if "timeout" == reason {
			retryAfter = 1 * time.Second
		}
		err = f.cluster.GetStorage().Retry(wfName, flowId, actionId, retryAfter)
		if err != nil {
			logger.Error("error retrying workflow", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
		}
	} else {
		logger.Error("action max retry exhausted, failing workflow", zap.Int("maxRetry", actionDefinition.RetryCount))
		f.MarkFailed(wfName, flowId)
	}

}

func (f *FlowEngine) DelayAction(wfName string, flowId string, actionId int, tryCount int, delay time.Duration) {
	err := f.validateFlow(wfName, flowId, actionId)
	if err != nil {
		return
	}
	err = f.cluster.GetStorage().Delay(wfName, flowId, actionId, delay)

	if err != nil {
		logger.Error("error running delay action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
	}
}

func (f *FlowEngine) ExecuteRetry(wfName string, flowId string, actionId int) {
	req := model.FlowExecutionRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		ActionId:     actionId,
		RequestType:  model.RETRY_FLOW_EXECUTION,
	}
	f.executionChannel <- req
}

func (f *FlowEngine) ExecuteResume(wfName string, flowId string, event string) {
	req := model.FlowExecutionRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		Event:        event,
		RequestType:  model.RESUME_FLOW_EXECUTION,
	}
	f.executionChannel <- req
}

func (f *FlowEngine) retry(wfName string, flowId string, actionId int) {
	flow, flowCtx, err := f.validateAndGetFlow(wfName, flowId, actionId)
	if err != nil {
		return
	}
	flowCtx.CurrentActionIds[actionId] = flowCtx.CurrentActionIds[actionId] + 1
	err = f.saveContextAndDispatchAction(wfName, flowId, []int{actionId}, flow, flowCtx)
	if err != nil {
		logger.Error("error dispatching action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
	}
}

func (f *FlowEngine) resume(wfName string, flowId string) {
	flowCtx, err := f.cluster.GetStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed, can not dispatch next action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	if flowCtx.State == model.COMPLETED || flowCtx.State == model.FAILED {
		logger.Debug("flow already completed, can not dispatch next action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	for k := range flowCtx.CurrentActionIds {
		req := model.FlowExecutionRequest{
			WorkflowName: wfName,
			FlowId:       flowId,
			Event:        "default",
			ActionId:     k,
			RequestType:  model.NEW_FLOW_EXECUTION,
		}
		f.execute(req)
	}
}

func (f *FlowEngine) MarkFailed(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.FAILED,
	}
}

func (f *FlowEngine) MarkPaused(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.PAUSED,
	}
}

func (f *FlowEngine) MarkWaitingDelay(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.WAITING_DELAY,
	}
	logger.Info("workflow waiting delay", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowEngine) MarkWaitingEvent(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.WAITING_EVENT,
	}
	logger.Info("workflow waiting for event", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowEngine) MarkRunning(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.RUNNING,
	}
	logger.Info("workflow running", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowEngine) changeState(wfName string, flowId string, state model.FlowState) {
	flowCtx, err := f.cluster.GetStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed", zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	flowCtx.State = state
	f.cluster.GetStorage().SaveFlowContext(wfName, flowId, flowCtx)
	if state == model.FAILED {
		flow, err := f.metadataService.GetFlow(wfName, flowId)
		if err != nil {
			logger.Error("Workflow not found", zap.String("Workflow", wfName), zap.Error(err))
			return
		}
		failureHandler := f.cluster.GetStateHandler().GetHandler(flow.FailureHandler)
		err = failureHandler(wfName, flowId)
		if err != nil {
			logger.Error("error in running failure handler", zap.Error(err))
		}
		logger.Info("workflow failed", zap.String("workflow", wfName), zap.String("FlowId", flowId))
	} else if state == model.COMPLETED {
		flow, err := f.metadataService.GetFlow(wfName, flowId)
		if err != nil {
			logger.Error("Workflow not found", zap.String("Workflow", wfName), zap.Error(err))
			return
		}
		successHandler := f.cluster.GetStateHandler().GetHandler(flow.SuccessHandler)
		err = successHandler(wfName, flowId)
		if err != nil {
			logger.Error("error in running success handler", zap.Error(err))
		}
		logger.Info("workflow completed", zap.String("workflow", wfName), zap.String("FlowId", flowId))
	}
}
