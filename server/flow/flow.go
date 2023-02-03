package flow

import (
	"fmt"
	"strings"

	"github.com/mohitkumar/orchy/server/action"
)

type Statehandler string

const DELETE Statehandler = "DELETE"
const NOOP Statehandler = "NOOP"

func ValidateStateHandler(st string) error {

	if len(st) == 0 || strings.EqualFold(st, "DELETE") || strings.EqualFold(st, "NOOP") {
		return nil
	}
	return fmt.Errorf("invalid state handler %s", st)
}

type Flow struct {
	Id             string
	RootAction     int
	Actions        map[int]action.Action
	FailureHandler Statehandler
	SuccessHandler Statehandler
}
