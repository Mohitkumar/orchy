package flow

import (
	"fmt"
	"strings"

	"github.com/mohitkumar/orchy/server/action"
	"github.com/mohitkumar/orchy/server/cluster"
	"github.com/mohitkumar/orchy/server/metadata"
	"github.com/mohitkumar/orchy/server/model"
	v8 "rogchap.com/v8go"
)

type Flow struct {
	Id             string
	RootAction     int
	Actions        map[int]action.Action
	FailureHandler cluster.Statehandler
	SuccessHandler cluster.Statehandler
}

func Convert(wf *model.Workflow, id string, jsvm *v8.Isolate, metadataStorage metadata.MetadataStorage) *Flow {
	actionMap := make(map[int]action.Action)
	for _, actionDef := range wf.Actions {
		var flAct action.Action
		actionType := action.ToActionType(actionDef.Type)
		nextMap := actionDef.Next
		baseAction := action.NewBaseAction(actionDef.Id, actionType,
			actionDef.Name, actionDef.InputParams, nextMap, metadataStorage)
		flAct = baseAction
		if actionType == action.ACTION_TYPE_SYSTEM {
			if strings.EqualFold(actionDef.Name, "switch") {
				flAct = action.NewSwitchAction(actionDef.Expression, *baseAction)
			} else if strings.EqualFold(actionDef.Name, "delay") {
				flAct = action.NewDelayAction(actionDef.DelaySeconds, *baseAction)
			} else if strings.EqualFold(actionDef.Name, "wait") {
				flAct = action.NewWaitAction(actionDef.Event, *baseAction)
			} else if strings.EqualFold(actionDef.Name, "javascript") {
				flAct = action.NewJsAction(actionDef.Expression, *baseAction, jsvm)
			}
		} else {
			flAct = action.NewUserAction(*baseAction)
		}
		actionMap[actionDef.Id] = flAct
	}
	var stateHandlerFailure cluster.Statehandler
	var stateHandlerSuccess cluster.Statehandler
	if len(wf.OnFailure) > 0 {
		stateHandlerFailure = cluster.Statehandler(wf.OnFailure)
	} else {
		stateHandlerFailure = cluster.NOOP
	}
	if len(wf.OnSuccess) > 0 {
		stateHandlerSuccess = cluster.Statehandler(wf.OnSuccess)
	} else {
		stateHandlerSuccess = cluster.NOOP
	}

	flow := &Flow{
		Id:             id,
		RootAction:     wf.RootAction,
		Actions:        actionMap,
		FailureHandler: stateHandlerFailure,
		SuccessHandler: stateHandlerSuccess,
	}
	return flow
}

func Validate(wf *model.Workflow, metadataStorage metadata.MetadataStorage) error {
	err := cluster.ValidateStateHandler(wf.OnFailure)
	if err != nil {
		return err
	}
	err = cluster.ValidateStateHandler(wf.OnSuccess)
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
	fl := Convert(wf, "validate", nil, metadataStorage)
	for _, act := range fl.Actions {
		err := act.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}
