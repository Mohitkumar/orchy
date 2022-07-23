package flow

import (
	"strings"

	"github.com/mohitkumar/orchy/server/action"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/model"
)

type Flow struct {
	Id         string
	RootAction int
	Actions    map[int]action.Action
}

func Convert(wf *model.Workflow, id string, container *container.DIContiner) Flow {
	actionMap := make(map[int]action.Action)
	for _, actionDef := range wf.Actions {
		actionType := action.ToActionType(actionDef.Type)
		var flAct action.Action = action.NewBaseAction(actionDef.Id, actionType,
			actionDef.Name, actionDef.InputParams, container)
		if actionType == action.ACTION_TYPE_SYSTEM {
			if strings.EqualFold(actionDef.Name, "switch") {
				flAct = action.NewSwitchAction(actionDef.Id, actionType,
					actionDef.Name, actionDef.Expression, actionDef.Cases, container)
			} else if strings.EqualFold(actionDef.Name, "delay") {
				flAct = action.NewDelayAction(actionDef.Id, actionType,
					actionDef.Name, actionDef.DelaySeconds, actionDef.Next, container)
			}
		} else {
			flAct = action.NewUserAction(actionDef.Id, actionType,
				actionDef.Name, actionDef.InputParams, actionDef.Next, container)
		}
		actionMap[actionDef.Id] = flAct
	}
	flow := Flow{
		Id:         id,
		RootAction: wf.RootAction,
		Actions:    actionMap,
	}
	return flow
}
