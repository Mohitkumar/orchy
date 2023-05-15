package shard

import (
	"fmt"
	"time"

	"github.com/mohitkumar/orchy/action"
	"github.com/mohitkumar/orchy/analytics"
	"github.com/mohitkumar/orchy/flow"
	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/metadata"
	"github.com/mohitkumar/orchy/model"
	"go.uber.org/zap"
)

type FlowEngine struct {
	storage               Storage
	metadataService       metadata.MetadataService
	stateHandler          *StateHandlerContainer
	stateMachineContainer *FlowStateMachineContainer
}

func NewFlowEngine(storage Storage, metadadataService metadata.MetadataService, stateMachineContainer *FlowStateMachineContainer) *FlowEngine {
	return &FlowEngine{
		storage:               storage,
		stateHandler:          NewStateHandlerContainer(storage),
		metadataService:       metadadataService,
		stateMachineContainer: stateMachineContainer,
	}
}

func (f *FlowEngine) GetFlow(wfName string, flowId string) (*model.FlowContext, error) {
	return f.storage.GetFlowContext(wfName, flowId)
}

func (f *FlowEngine) ExecuteAction(wfName string, wfId string, event string, actionId int, data map[string]any) {
	f.execute(wfName, wfId, actionId, event, data)
}

func (f *FlowEngine) Init(wfName string, flowId string, input map[string]any) error {
	stateMachine, err := f.stateMachineContainer.Init(wfName, flowId, input)
	if err != nil {
		return err
	}
	f.saveContextAndDispatchAction(wfName, flowId, []int{stateMachine.flow.RootAction}, stateMachine.flow, stateMachine.context)
	if err != nil {
		logger.Error("error executing flow", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(err))
		return err
	}
	return nil
}

func (f *FlowEngine) execute(wfName string, flowId string, actionId int, event string, dataMap map[string]any) {
	stateMachine, err := f.stateMachineContainer.Get(wfName, flowId)
	if err != nil {
		return
	}
	err = stateMachine.Validate(actionId)
	if err != nil {
		return
	}
	complete, actionIdsToDispatch := stateMachine.MoveForward(actionId, event, dataMap)
	if complete {
		f.MarkComplete(wfName, flowId, stateMachine.flow.SuccessHandler)
		return
	}
	stateMachine.context.State = model.RUNNING
	err = f.saveContextAndDispatchAction(wfName, flowId, actionIdsToDispatch, stateMachine.flow, stateMachine.context)
	if err != nil {
		logger.Error("error executiong flow", zap.Any("Workflow", wfName), zap.String("flowId", flowId), zap.Int("action", actionId), zap.Error(err))
	}
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
	err := f.storage.SaveFlowContextAndDispatchAction(wfName, flowId, flowCtx, actions)
	if err != nil {
		return err
	}
	return nil
}

func (f *FlowEngine) ExecuteSystemAction(wfName string, flowId string, actionId int) {
	stateMachine, err := f.stateMachineContainer.Get(wfName, flowId)
	if err != nil {
		return
	}
	err = stateMachine.Validate(actionId)
	if err != nil {
		return
	}
	flow := stateMachine.flow
	flowCtx := stateMachine.context
	currentAction := flow.Actions[actionId]
	event, dataMap, err := currentAction.Execute(wfName, flowCtx, 1)
	if err != nil {
		logger.Error("error executing workflow", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId), zap.Error(err))
		analytics.RecordActionFailure(wfName, flowId, currentAction.GetName(), actionId, "action failed")
		f.MarkFailed(wfName, flowId)
		return
	}

	if currentAction.GetType() == action.ACTION_TYPE_SYSTEM {
		switch currentAction.GetName() {
		case "delay":
			f.MarkWaitingDelay(wfName, flowId)
			delay := currentAction.GetParams()["delay"].(time.Duration)
			f.DelayAction(wfName, flowId, currentAction.GetName(), currentAction.GetId(), delay)
		case "wait":
			timeout := currentAction.GetParams()["timeout"].(time.Duration)
			event := currentAction.GetParams()["event"].(string)
			if timeout != 0 {
				f.DelayAction(wfName, flowId, currentAction.GetName(), currentAction.GetId(), timeout)
			}
			f.MarkWaitingEvent(wfName, flowId, event)
		default:
			f.ExecuteAction(wfName, flowId, event, actionId, dataMap)
		}
		analytics.RecordActionSuccess(wfName, flowId, currentAction.GetName(), actionId, dataMap)
	} else {
		logger.Error("should be system action")
	}
}

func (f *FlowEngine) RetryFailedAction(wfName string, flowId string, actionName string, actionId int) {
	stateMachine, err := f.stateMachineContainer.Get(wfName, flowId)
	if err != nil {
		return
	}
	err = stateMachine.Validate(actionId)
	if err != nil {
		return
	}
	flow := stateMachine.flow
	flowCtx := stateMachine.context

	action := flow.Actions[actionId]
	actionDefinition, err := f.metadataService.GetMetadataStorage().GetActionDefinition(action.GetName())
	if err != nil {
		logger.Error("action definition not found ", zap.String("action", action.GetName()), zap.Error(err))
		return
	}
	tryCount := flowCtx.CurrentActionIds[actionId]
	logger.Info("retrying action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.String("reason", "failed"), zap.Int("action", actionId), zap.Int("retry", tryCount+1))
	if tryCount < actionDefinition.RetryCount {
		var retryAfter time.Duration
		switch actionDefinition.RetryPolicy {
		case model.RETRY_POLICY_FIXED:
			retryAfter = time.Duration(actionDefinition.RetryAfterSeconds) * time.Second
		case model.RETRY_POLICY_BACKOFF:
			retryAfter = time.Duration(actionDefinition.RetryAfterSeconds*(tryCount+1)) * time.Second
		}
		err = f.storage.Retry(wfName, flowId, actionName, actionId, retryAfter)
		if err != nil {
			logger.Error("error retrying workflow", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
		}
	} else {
		logger.Error("action max retry exhausted, failing workflow", zap.Int("maxRetry", actionDefinition.RetryCount))
		f.MarkFailed(wfName, flowId)
	}

}

func (f *FlowEngine) RetryTimedoutAction(wfName string, flowId string, actionName string, actionId int) {
	stateMachine, err := f.stateMachineContainer.Get(wfName, flowId)
	if err != nil {
		return
	}
	err = stateMachine.Validate(actionId)
	if err != nil {
		return
	}
	flow := stateMachine.flow
	flowCtx := stateMachine.context

	action := flow.Actions[actionId]
	actionDefinition, err := f.metadataService.GetMetadataStorage().GetActionDefinition(action.GetName())
	if err != nil {
		logger.Error("action definition not found ", zap.String("action", action.GetName()), zap.Error(err))
		return
	}
	tryCount := flowCtx.CurrentActionIds[actionId]
	logger.Info("retrying action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.String("reason", "tinedout"), zap.Int("action", actionId), zap.Int("retry", tryCount+1))
	if tryCount < actionDefinition.RetryCount {
		err = f.storage.Retry(wfName, flowId, actionName, actionId, 1*time.Second)
		if err != nil {
			logger.Error("error retrying workflow", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
		}
	} else {
		analytics.RecordActionFailure(wfName, flowId, actionName, actionId, "action timedout")
		logger.Error("action max retry exhausted, failing workflow", zap.Int("maxRetry", actionDefinition.RetryCount))
		f.MarkFailed(wfName, flowId)
	}

}

func (f *FlowEngine) DelayAction(wfName string, flowId string, actionName string, actionId int, delay time.Duration) {
	stateMachine, err := f.stateMachineContainer.Get(wfName, flowId)
	if err != nil {
		return
	}
	err = stateMachine.Validate(actionId)
	if err != nil {
		return
	}

	err = f.storage.Delay(wfName, flowId, actionName, actionId, delay)

	if err != nil {
		logger.Error("error running delay action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
	}
}

func (f *FlowEngine) ExecuteDelay(wfName string, flowId string, actionId int) {
	f.ExecuteAction(wfName, flowId, "default", actionId, nil)
}

func (f *FlowEngine) ExecuteRetry(wfName string, flowId string, actionId int) {
	stateMachine, err := f.stateMachineContainer.Get(wfName, flowId)
	if err != nil {
		return
	}
	err = stateMachine.Validate(actionId)
	if err != nil {
		return
	}
	flow := stateMachine.flow
	flowCtx := stateMachine.context
	flowCtx.CurrentActionIds[actionId] = flowCtx.CurrentActionIds[actionId] + 1
	err = f.saveContextAndDispatchAction(wfName, flowId, []int{actionId}, flow, flowCtx)
	if err != nil {
		logger.Error("error dispatching action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Int("action", actionId))
	}
}

func (f *FlowEngine) ExecuteResume(wfName string, flowId string, event string) {
	stateMachine, err := f.stateMachineContainer.Get(wfName, flowId)
	if err != nil {
		return
	}
	flowCtx := stateMachine.context
	if flowCtx.State == model.COMPLETED || flowCtx.State == model.FAILED {
		logger.Debug("flow already completed, can not dispatch next action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(fmt.Errorf("workflow complted")))
		return
	}
	for k := range flowCtx.CurrentActionIds {
		f.ExecuteAction(wfName, flowId, event, k, nil)
	}
}

func (f *FlowEngine) ExecuteResumeAfterWait(wfName string, flowId string, event string) error {
	stateMachine, err := f.stateMachineContainer.Get(wfName, flowId)
	if err != nil {
		return err
	}
	flowCtx := stateMachine.context
	if flowCtx.Event != event {
		logger.Error("Workflow event mismatch", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.String("Event Received", event), zap.String("Workflow Event", flowCtx.Event))
		return fmt.Errorf("can not consume event %s, workflow waiting on event %s", event, flowCtx.Event)
	}
	if flowCtx.State == model.COMPLETED || flowCtx.State == model.FAILED {
		logger.Debug("flow already completed, can not dispatch next action", zap.String("Workflow", wfName), zap.String("FlowId", flowId), zap.Error(fmt.Errorf("workflow complted")))
		return fmt.Errorf("flow already completed")
	}
	for k := range flowCtx.CurrentActionIds {
		f.ExecuteAction(wfName, flowId, event, k, nil)
	}
	return nil
}

func (f *FlowEngine) MarkComplete(wfName string, flowId string, successhandler flow.Statehandler) {
	f.changeState(wfName, flowId, model.COMPLETED)
}

func (f *FlowEngine) MarkFailed(wfName string, flowId string) {
	f.changeState(wfName, flowId, model.FAILED)
}

func (f *FlowEngine) MarkPaused(wfName string, flowId string) {
	f.changeState(wfName, flowId, model.PAUSED)
}

func (f *FlowEngine) MarkWaitingDelay(wfName string, flowId string) {
	f.changeState(wfName, flowId, model.WAITING_DELAY)
	logger.Info("workflow waiting delay", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowEngine) MarkWaitingEvent(wfName string, flowId string, event string) {
	f.changeStateWithEvent(wfName, flowId, model.WAITING_EVENT, event)
	logger.Info("workflow waiting for event", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowEngine) MarkRunning(wfName string, flowId string) {
	f.changeState(wfName, flowId, model.RUNNING)
	logger.Info("workflow running", zap.String("workflow", wfName), zap.String("FlowId", flowId))
}

func (f *FlowEngine) changeState(wfName string, flowId string, state model.FlowState) {
	f.changeStateWithEvent(wfName, flowId, state, "")
}

func (f *FlowEngine) changeStateWithEvent(wfName string, flowId string, state model.FlowState, event string) {
	stateMachine, err := f.stateMachineContainer.Get(wfName, flowId)
	if err != nil {
		return
	}
	stateMachine.ChangeState(state)
	if len(event) != 0 {
		stateMachine.SaveEvent(event)
	}
	flow := stateMachine.flow
	f.storage.SaveFlowContext(wfName, flowId, stateMachine.context)
	if state != model.RUNNING {
		f.stateMachineContainer.Delete(wfName, flowId)
	}
	if state == model.FAILED {
		failureHandler := f.stateHandler.GetHandler(flow.FailureHandler)
		err = failureHandler(wfName, flowId)
		if err != nil {
			logger.Error("error in running failure handler", zap.Error(err))
		}
		logger.Info("workflow failed", zap.String("workflow", wfName), zap.String("FlowId", flowId))
	} else if state == model.COMPLETED {
		successHandler := f.stateHandler.GetHandler(flow.SuccessHandler)
		err = successHandler(wfName, flowId)
		if err != nil {
			logger.Error("error in running success handler", zap.Error(err))
		}
		logger.Info("workflow completed", zap.String("workflow", wfName), zap.String("FlowId", flowId))
	}
}
