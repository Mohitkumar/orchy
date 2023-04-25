package executor

import (
	"sync"

	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/shard"
	"github.com/mohitkumar/orchy/util"
	"go.uber.org/zap"
)

var _ shard.Executor = new(systemActionExecutor)

type timeoutExecutor struct {
	shardId string
	storage shard.Storage
	engine  *shard.FlowEngine
	wg      *sync.WaitGroup
	tw      *util.Worker
	stop    chan struct{}
}

func NewTimeoutExecutor(shardId string, storage shard.Storage, engine *shard.FlowEngine, wg *sync.WaitGroup) *timeoutExecutor {
	ex := &timeoutExecutor{
		shardId: shardId,
		storage: storage,
		engine:  engine,
		stop:    make(chan struct{}),
		wg:      wg,
	}
	ex.tw = util.NewWorker("timeout-executor-"+shardId, ex.stop, ex.handle, ex.wg)
	return ex
}

func (ex *timeoutExecutor) Start() {
	if ex.IsRunning() {
		return
	}
	ex.tw.Start()
}

func (ex *timeoutExecutor) Stop() {
	if !ex.IsRunning() {
		return
	}
	ex.stop <- struct{}{}
}

func (ex *timeoutExecutor) IsRunning() bool {
	return ex.tw.IsRunning()
}

func (ex *timeoutExecutor) handle() bool {
	actions, err := ex.storage.PollTimeout()
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
		return false
	}
	if len(actions) == 0 {
		return false
	}
	for _, action := range actions {
		ex.engine.RetryTimedoutAction(action.WorkflowName, action.FlowId, action.ActionName, action.ActionId)
	}
	return true
}
