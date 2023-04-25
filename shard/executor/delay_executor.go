package executor

import (
	"sync"

	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/model"
	"github.com/mohitkumar/orchy/shard"
	"github.com/mohitkumar/orchy/util"
	"go.uber.org/zap"
)

var _ shard.Executor = new(systemActionExecutor)

type delayExecutor struct {
	shardId          string
	storage          shard.Storage
	engine           *shard.FlowEngine
	wg               *sync.WaitGroup
	tw               *util.Worker
	executionChannel chan<- model.FlowExecutionRequest
	stop             chan struct{}
}

func NewDelayExecutor(shardId string, storage shard.Storage, engine *shard.FlowEngine, wg *sync.WaitGroup) *delayExecutor {
	ex := &delayExecutor{
		shardId: shardId,
		storage: storage,
		engine:  engine,
		stop:    make(chan struct{}),
		wg:      wg,
	}
	ex.tw = util.NewWorker("delay-executor-"+shardId, ex.stop, ex.handle, ex.wg)
	return ex
}

func (ex *delayExecutor) Start() {
	if ex.IsRunning() {
		return
	}
	ex.tw.Start()
}

func (ex *delayExecutor) IsRunning() bool {
	return ex.tw.IsRunning()
}

func (ex *delayExecutor) Stop() {
	if !ex.IsRunning() {
		return
	}
	ex.stop <- struct{}{}
}

func (ex *delayExecutor) handle() bool {
	actions, err := ex.storage.PollDelay()
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
		return false
	}
	if len(actions) == 0 {
		return false
	}
	for _, action := range actions {
		ex.engine.ExecuteDelay(action.WorkflowName, action.FlowId, action.ActionId)
	}
	return true
}
