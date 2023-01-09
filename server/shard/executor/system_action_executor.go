package executor

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/shard"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ shard.Executor = new(systemActionExecutor)

type systemActionExecutor struct {
	shardId          string
	storage          shard.Storage
	executionChannel chan<- model.FlowExecutionRequest
	tw               *util.TickWorker
	wg               *sync.WaitGroup
	stop             chan struct{}
}

func NewSystemActionExecutor(shardId string, storage shard.Storage, executionChannel chan<- model.FlowExecutionRequest, wg *sync.WaitGroup) *systemActionExecutor {
	ex := &systemActionExecutor{
		shardId:          shardId,
		storage:          storage,
		executionChannel: executionChannel,
		stop:             make(chan struct{}),
		wg:               wg,
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
	actions, err := ex.storage.PollAction("system", 10)
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions {
		ex.flowService.ExecuteSystemAction(action.WorkflowName, action.FlowId, 1, action.ActionId)
	}
}
