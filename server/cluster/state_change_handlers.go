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
	flowDao  FlowDao
}

func NewStateHandlerContainer(flowDao FlowDao) *StateHandlerContainer {
	hd := &StateHandlerContainer{
		flowDao:  flowDao,
		handlers: make(map[Statehandler]func(wfName string, wfId string) error, 1),
	}
	return hd
}

func (s *StateHandlerContainer) Init() {
	s.handlers[DELETE] = s.delete
}

func (s *StateHandlerContainer) GetHandler(st Statehandler) func(wfName string, wfId string) error {
	handler, ok := s.handlers[st]
	if ok {
		return handler
	}
	return s.noop
}
func (s *StateHandlerContainer) delete(wfName string, wfId string) error {
	return s.flowDao.DeleteFlowContext(wfName, wfId)
}

func (s *StateHandlerContainer) noop(wfName string, wfId string) error {
	logger.Info("noop handler called")
	return nil
}
