package cluster

import (
	"fmt"
	"strings"

	"github.com/mohitkumar/orchy/server/logger"
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

type StateHandlerContainer struct {
	handlers map[Statehandler]func(wfName string, wfId string) error
	storage  Storage
}

func NewStateHandlerContainer(storage Storage) *StateHandlerContainer {
	hd := &StateHandlerContainer{
		storage:  storage,
		handlers: make(map[Statehandler]func(wfName string, wfId string) error, 1),
	}
	hd.handlers[DELETE] = hd.delete
	hd.handlers[NOOP] = hd.noop
	return hd
}

func (s *StateHandlerContainer) GetHandler(st Statehandler) func(wfName string, wfId string) error {
	handler, ok := s.handlers[st]
	if ok {
		return handler
	}
	return s.noop
}
func (s *StateHandlerContainer) delete(wfName string, wfId string) error {
	return s.storage.DeleteFlowContext(wfName, wfId)
}

func (s *StateHandlerContainer) noop(wfName string, wfId string) error {
	logger.Info("noop handler called")
	return nil
}
