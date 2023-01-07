package flow

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mohitkumar/orchy/server/action"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
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

func (f *FlowService) ExecuteAction(wfName string, wfId string, event string, actionId int, data map[string]any) {
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
					f.retry(req.WorkflowName, req.FlowId, req.ActionId)
				case model.RESUME_FLOW_EXECUTION:
					f.resume(req.WorkflowName, req.FlowId)
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
		ExecutedActions:  map[int]bool{},
		CurrentActionIds: map[int]int{wf.RootAction: 1},
	}
	f.saveContextAndDispatchAction(wfName, flowId, []int{wf.RootAction}, flow, flowCtx)
	if err != nil {
		logger.Error("error executiong flow", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(err))
		return flowId, err
	}
	return flowId, nil
}

func (f *FlowService) execute(req model.FlowExecutionRequest) {
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

func (f *FlowService) validateFlow(wfName string, flowId string, actionId int) error {
	flowCtx, err := f.container.GetClusterStorage().GetFlowContext(wfName, flowId)
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

func (f *FlowService) validateAndGetFlow(wfName string, flowId string, actionId int) (*Flow, *model.FlowContext, error) {
	wf, err := f.container.GetMetadataStorage().GetWorkflowDefinition(wfName)
	if err != nil {
		logger.Error("Workflow Definition not found", zap.String("Workflow", wfName), zap.Error(err))
		return nil, nil, err
	}
	flow := Convert(wf, flowId, f.container)
	flowCtx, err := f.container.GetClusterStorage().GetFlowContext(wfName, flowId)
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
func (f *FlowService) saveContextAndDispatchAction(wfName string, flowId string, actionIds []int, flow *Flow, flowCtx *model.FlowContext) error {
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
	flow, flowCtx, err := f.validateAndGetFlow(wfName, flowId, actionId)
	if err != nil {
		return
	}
	currentAction := flow.Actions[actionId]
	event, dataMap, err := currentAction.Execute(wfName, flowCtx, tryCount)
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

func (f *FlowService) RetryAction(wfName string, flowId string, actionId int, reason string) {
	flow, flowCtx, err := f.validateAndGetFlow(wfName, flowId, actionId)
	if err != nil {
		return
	}
	action := flow.Actions[actionId]
	actionDefinition, err := f.container.GetMetadataStorage().GetActionDefinition(action.GetName())
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
		err = f.container.GetClusterStorage().Retry(wfName, flowId, actionId, retryAfter)
		if err != nil {
			logger.Error("error retrying workflow", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
		}
	} else {
		logger.Error("action max retry exhausted, failing workflow", zap.Int("maxRetry", actionDefinition.RetryCount))
		f.MarkFailed(wfName, flowId)
	}

}

func (f *FlowService) DelayAction(wfName string, flowId string, actionId int, tryCount int, delay time.Duration) {
	err := f.validateFlow(wfName, flowId, actionId)
	if err != nil {
		return
	}
	err = f.container.GetClusterStorage().Delay(wfName, flowId, actionId, delay)

	if err != nil {
		logger.Error("error running delay action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
	}
}

func (f *FlowService) ExecuteRetry(wfName string, flowId string, actionId int) {
	req := model.FlowExecutionRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		ActionId:     actionId,
		RequestType:  model.RETRY_FLOW_EXECUTION,
	}
	f.executionChannel <- req
}

func (f *FlowService) ExecuteResume(wfName string, flowId string, event string) {
	req := model.FlowExecutionRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		Event:        event,
		RequestType:  model.RESUME_FLOW_EXECUTION,
	}
	f.executionChannel <- req
}

func (f *FlowService) retry(wfName string, flowId string, actionId int) {
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

func (f *FlowService) resume(wfName string, flowId string) {
	flowCtx, err := f.container.GetClusterStorage().GetFlowContext(wfName, flowId)
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

func (f *FlowService) MarkFailed(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.FAILED,
	}
}

func (f *FlowService) MarkPaused(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.PAUSED,
	}
}

func (f *FlowService) MarkWaitingDelay(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.WAITING_DELAY,
	}
	logger.Info("workflow waiting delay", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowService) MarkWaitingEvent(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.WAITING_EVENT,
	}
	logger.Info("workflow waiting for event", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowService) MarkRunning(wfName string, flowId string) {
	f.stateChangeChannel <- model.FlowStateChangeRequest{
		WorkflowName: wfName,
		FlowId:       flowId,
		State:        model.RUNNING,
	}
	logger.Info("workflow running", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowService) changeState(wfName string, flowId string, state model.FlowState) {
	flowCtx, err := f.container.GetClusterStorage().GetFlowContext(wfName, flowId)
	if err != nil {
		logger.Debug("flow already completed", zap.Error(fmt.Errorf("workflow complted")))
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
