package shard

import (
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/model"
)

type ExternalQueue interface {
	Push(action *api.Action) error
	Poll(actionName string, batchSize int) (*api.Actions, error)
}

type Storage interface {
	SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error
	GetFlowContext(wfName string, flowId string) (*model.FlowContext, error)
	DeleteFlowContext(wfName string, flowId string) error

	SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, actions []model.ActionExecutionRequest) error
	PollAction(actionType string, batchSize int) ([]model.ActionExecutionRequest, error)
	Ack(actionType string, actions []model.ActionExecutionRequest) error
	Retry(wfName string, flowId string, actionName string, actionId int, delay time.Duration) error
	PollRetry() ([]model.ActionExecutionRequest, error)
	Delay(wfName string, flowId string, actionName string, actionId int, delay time.Duration) error
	PollDelay() ([]model.ActionExecutionRequest, error)
	Timeout(wfName string, flowId string, actionName string, actionId int, delay time.Duration) error
	PollTimeout() ([]model.ActionExecutionRequest, error)
}

type Executor interface {
	Start()
	Stop()
}

type Shard struct {
	id            string
	externalQueue ExternalQueue
	storage       Storage
	stateHandler  *StateHandlerContainer
	engine        *FlowEngine
	executors     map[string]Executor
}

func NewShard(id string, externalQueue ExternalQueue, storage Storage, engine *FlowEngine, stateHandler *StateHandlerContainer) *Shard {
	return &Shard{
		id:            id,
		externalQueue: externalQueue,
		storage:       storage,
		engine:        engine,
		stateHandler:  stateHandler,
		executors:     make(map[string]Executor),
	}
}

func (s *Shard) RegisterExecutor(name string, executor Executor) {
	s.executors[name] = executor
}

func (s *Shard) GetShardId() string {
	return s.id
}

func (s *Shard) GetStorage() Storage {
	return s.storage
}

func (s *Shard) GetEngine() *FlowEngine {
	return s.engine
}

func (s *Shard) GetExternalQueue() ExternalQueue {
	return s.externalQueue
}

func (s *Shard) Start() {
	for _, executor := range s.executors {
		executor.Start()
	}
}

func (s *Shard) Stop() {
	for _, executor := range s.executors {
		executor.Stop()
	}
}
