package flow

import (
	"fmt"
	"strings"

	"github.com/mohitkumar/orchy/server/action"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
)

type Flow struct {
	Id             string
	RootAction     int
	Actions        map[int]action.Action
	FailureHandler persistence.Statehandler
	SuccessHandler persistence.Statehandler
}

func Convert(wf *model.Workflow, id string, container *container.DIContiner) *Flow {
	actionMap := make(map[int]action.Action)
	for _, actionDef := range wf.Actions {
		var flAct action.Action
		actionType := action.ToActionType(actionDef.Type)
		nextMap := actionDef.Next
		baseAction := action.NewBaseAction(actionDef.Id, actionType,
			actionDef.Name, actionDef.InputParams, nextMap, container)
		flAct = baseAction
		if actionType == action.ACTION_TYPE_SYSTEM {
			if strings.EqualFold(actionDef.Name, "switch") {
				flAct = action.NewSwitchAction(actionDef.Expression, *baseAction)
			} else if strings.EqualFold(actionDef.Name, "delay") {
				flAct = action.NewDelayAction(actionDef.DelaySeconds, *baseAction)
			}
		} else {
			flAct = action.NewUserAction(*baseAction)
		}
		actionMap[actionDef.Id] = flAct
	}
	var stateHandlerFailure persistence.Statehandler
	var stateHandlerSuccess persistence.Statehandler
	if len(wf.OnFailure) > 0 {
		stateHandlerFailure = persistence.Statehandler(wf.OnFailure)
	} else {
		stateHandlerFailure = persistence.NOOP
	}
	if len(wf.OnSuccess) > 0 {
		stateHandlerSuccess = persistence.Statehandler(wf.OnSuccess)
	} else {
		stateHandlerSuccess = persistence.NOOP
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

func Validate(wf *model.Workflow, container *container.DIContiner) error {
	err := persistence.ValidateStateHandler(wf.OnFailure)
	if err != nil {
		return err
	}
	err = persistence.ValidateStateHandler(wf.OnSuccess)
	if err != nil {
		return err
	}
	validActionId := make(map[int]any)
	for _, actionDef := range wf.Actions {
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
	fl := Convert(wf, "validate", container)
	for _, act := range fl.Actions {
		err := act.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}
