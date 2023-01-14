package executor

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/shard"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ shard.Executor = new(systemActionExecutor)

type retryExecutor struct {
	shardId string
	storage shard.Storage
	engine  *shard.FlowEngine
	tw      *util.TickWorker
	wg      *sync.WaitGroup
	stop    chan struct{}
}

func NewRetryExecutor(shardId string, stoarge shard.Storage, engine *shard.FlowEngine, wg *sync.WaitGroup) *retryExecutor {
	ex := &retryExecutor{
		shardId: shardId,
		storage: stoarge,
		engine:  engine,
		stop:    make(chan struct{}),
		wg:      wg,
	}
	ex.tw = util.NewTickWorker("retryexecutor-"+shardId, 1*time.Second, ex.stop, ex.handle, ex.wg)
	return ex
}

func (ex *retryExecutor) Start() {
	if ex.IsRunning() {
		return
	}
	ex.tw.Start()
}

func (ex *retryExecutor) IsRunning() bool {
	return ex.tw.IsRunning()
}

func (ex *retryExecutor) Stop() {
	if !ex.IsRunning() {
		return
	}
	ex.stop <- struct{}{}
}

func (ex *retryExecutor) handle() {
	actions, err := ex.storage.PollRetry()
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions {
		ex.engine.ExecuteRetry(action.WorkflowName, action.FlowId, action.ActionId)
	}
}
