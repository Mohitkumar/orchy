package executor

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/shard"
	"github.com/mohitkumar/orchy/util"
	"go.uber.org/zap"
)

var _ shard.Executor = new(systemActionExecutor)

type systemActionExecutor struct {
	shardId   string
	storage   shard.Storage
	engine    *shard.FlowEngine
	tw        *util.TickWorker
	wg        *sync.WaitGroup
	batchSize int
	stop      chan struct{}
}

func NewSystemActionExecutor(shardId string, storage shard.Storage, engine *shard.FlowEngine, batchSize int, wg *sync.WaitGroup) *systemActionExecutor {
	ex := &systemActionExecutor{
		shardId:   shardId,
		storage:   storage,
		engine:    engine,
		stop:      make(chan struct{}),
		batchSize: batchSize,
		wg:        wg,
	}
	ex.tw = util.NewTickWorker("system-action-executor-"+shardId, 1*time.Second, ex.stop, ex.handle, ex.wg)
	return ex
}

func (ex *systemActionExecutor) Start() {
	if ex.IsRunning() {
		return
	}
	ex.tw.Start()
}

func (ex *systemActionExecutor) Stop() {
	if !ex.IsRunning() {
		return
	}
	ex.stop <- struct{}{}
}

func (ex *systemActionExecutor) IsRunning() bool {
	return ex.tw.IsRunning()
}

func (ex *systemActionExecutor) handle() {
	actions, err := ex.storage.PollAction("system", ex.batchSize)
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions {
		ex.engine.ExecuteSystemAction(action.WorkflowName, action.FlowId, action.ActionId)
	}
	if len(actions) != 0 {
		err = ex.storage.Ack("system", actions)
		if err != nil {
			logger.Error("error while ack system actions", zap.Error(err))
		}
	}
}
