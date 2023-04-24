package metadata

import (
	"fmt"
	"strings"

	"github.com/mohitkumar/orchy/action"
	"github.com/mohitkumar/orchy/flow"
	"github.com/mohitkumar/orchy/model"
	v8 "rogchap.com/v8go"
)

type MetadataService interface {
	GetFlow(name string, id string) (*flow.Flow, error)
	ValidateFlow(wf model.Workflow) error
	GetMetadataStorage() MetadataStorage
}

type MetadataServiceImpl struct {
	storage MetadataStorage
	jsVm    *v8.Isolate
}

func NewMetadataService(storage MetadataStorage, jsVm *v8.Isolate) MetadataService {
	return &MetadataServiceImpl{
		storage: storage,
		jsVm:    jsVm,
	}
}

func (s *MetadataServiceImpl) GetFlow(name string, id string) (*flow.Flow, error) {
	wf, err := s.storage.GetWorkflowDefinition(name)
	if err != nil {
		return nil, err
	}
	actionMap := make(map[int]action.Action)
	for _, actionDef := range wf.Actions {
		var flAct action.Action
		actionType := action.ToActionType(actionDef.Type)
		nextMap := actionDef.Next
		baseAction := action.NewBaseAction(actionDef.Id, actionType,
			actionDef.Name, actionDef.InputParams, nextMap)
		flAct = baseAction
		if actionType == action.ACTION_TYPE_SYSTEM {
			if strings.EqualFold(actionDef.Name, "switch") {
				flAct = action.NewSwitchAction(actionDef.Expression, *baseAction)
			} else if strings.EqualFold(actionDef.Name, "delay") {
				flAct = action.NewDelayAction(actionDef.DelaySeconds, *baseAction)
			} else if strings.EqualFold(actionDef.Name, "wait") {
				flAct = action.NewWaitAction(actionDef.Event, actionDef.TimeoutSeconds, *baseAction)
			} else if strings.EqualFold(actionDef.Name, "javascript") {
				flAct = action.NewJsAction(actionDef.Expression, *baseAction, s.jsVm)
			} else if strings.EqualFold(actionDef.Name, "jsonmapper") {
				flAct = action.NewJsonMapAction(*baseAction)
			}
		} else {
			flAct = action.NewUserAction(*baseAction)
		}
		actionMap[actionDef.Id] = flAct
	}
	var stateHandlerFailure flow.Statehandler
	var stateHandlerSuccess flow.Statehandler
	if len(wf.OnFailure) > 0 {
		stateHandlerFailure = flow.Statehandler(wf.OnFailure)
	} else {
		stateHandlerFailure = flow.NOOP
	}
	if len(wf.OnSuccess) > 0 {
		stateHandlerSuccess = flow.Statehandler(wf.OnSuccess)
	} else {
		stateHandlerSuccess = flow.NOOP
	}
	var terminalActions []int
	for _, act := range actionMap {
		if act.GetNext() == nil {
			terminalActions = append(terminalActions, act.GetId())
		}
	}
	flow := &flow.Flow{
		Id:              id,
		RootAction:      wf.RootAction,
		Actions:         actionMap,
		TerminalActions: terminalActions,
		FailureHandler:  stateHandlerFailure,
		SuccessHandler:  stateHandlerSuccess,
	}
	return flow, nil
}

func (s *MetadataServiceImpl) ValidateFlow(wf model.Workflow) error {
	err := flow.ValidateStateHandler(wf.OnFailure)
	if err != nil {
		return err
	}
	err = flow.ValidateStateHandler(wf.OnSuccess)
	if err != nil {
		return err
	}
	allActions := make(map[int]bool)
	for _, actionDef := range wf.Actions {
		allActions[actionDef.Id] = true
	}
	validActionId := make(map[int]any)
	for _, actionDef := range wf.Actions {
		nextMap := actionDef.Next
		if nextMap != nil {
			if len(nextMap) == 0 {
				return fmt.Errorf("invalid next action for action %d", actionDef.Id)
			}
			for _, v := range nextMap {
				if v == nil {
					return fmt.Errorf("invalid next action for action %d, should be array", actionDef.Id)
				}
				for _, act := range v {
					if _, ok := allActions[act]; !ok {
						return fmt.Errorf("invalid next action for action %d, action %d not defined", actionDef.Id, act)
					}
				}
			}
		}
		if _, ok := validActionId[actionDef.Id]; ok {
			return fmt.Errorf("action id %d is duplicate", actionDef.Id)
		}
		validActionId[actionDef.Id] = ""
		err := action.ValidateActionType(actionDef.Type)
		if err != nil {
			return err
		}
	}
	if _, ok := validActionId[wf.RootAction]; !ok {
		return fmt.Errorf("no action with root action id %d in workflow", wf.RootAction)
	}
	actionMap := make(map[int]action.Action)
	for _, actionDef := range wf.Actions {
		var flAct action.Action
		actionType := action.ToActionType(actionDef.Type)
		nextMap := actionDef.Next
		baseAction := action.NewBaseAction(actionDef.Id, actionType,
			actionDef.Name, actionDef.InputParams, nextMap)
		flAct = baseAction
		if actionType == action.ACTION_TYPE_SYSTEM {
			if strings.EqualFold(actionDef.Name, "switch") {
				flAct = action.NewSwitchAction(actionDef.Expression, *baseAction)
			} else if strings.EqualFold(actionDef.Name, "delay") {
				flAct = action.NewDelayAction(actionDef.DelaySeconds, *baseAction)
			} else if strings.EqualFold(actionDef.Name, "wait") {
				flAct = action.NewWaitAction(actionDef.Event, actionDef.TimeoutSeconds, *baseAction)
			} else if strings.EqualFold(actionDef.Name, "javascript") {
				flAct = action.NewJsAction(actionDef.Expression, *baseAction, s.jsVm)
			} else if strings.EqualFold(actionDef.Name, "jsonmapper") {
				flAct = action.NewJsonMapAction(*baseAction)
			}
		} else {
			flAct = action.NewUserAction(*baseAction)
		}
		actionMap[actionDef.Id] = flAct
	}
	for _, act := range actionMap {
		err := act.Validate()
		if err != nil {
			return err
		}
		if act.GetType() == action.ACTION_TYPE_USER {
			_, err := s.storage.GetActionDefinition(act.GetName())
			if err != nil {
				return fmt.Errorf("actionId=%d, action %s not registered", act.GetId(), act.GetName())
			}
		}
	}
	return nil
}

func (s *MetadataServiceImpl) GetMetadataStorage() MetadataStorage {
	return s.storage
}
