package shard

import (
	"github.com/mohitkumar/orchy/flow"
	"github.com/mohitkumar/orchy/logger"
)

type StateHandlerContainer struct {
	handlers map[flow.Statehandler]func(wfName string, wfId string) error
	storage  Storage
}

func NewStateHandlerContainer(storage Storage) *StateHandlerContainer {
	hd := &StateHandlerContainer{
		storage:  storage,
		handlers: make(map[flow.Statehandler]func(wfName string, wfId string) error, 1),
	}
	hd.handlers[flow.DELETE] = hd.delete
	hd.handlers[flow.NOOP] = hd.noop
	return hd
}

func (s *StateHandlerContainer) GetHandler(st flow.Statehandler) func(wfName string, wfId string) error {
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
