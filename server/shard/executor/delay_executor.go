package executor

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/shard"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ shard.Executor = new(systemActionExecutor)

type delayExecutor struct {
	flowService *flow.FlowService
	shardId     string
	storage     shard.Storage
	wg          *sync.WaitGroup
	tw          *util.TickWorker
	stop        chan struct{}
}

func NewDelayExecutor(shardId string, stoarge shard.Storage, flowService *flow.FlowService, wg *sync.WaitGroup) *delayExecutor {
	ex := &delayExecutor{
		flowService: flowService,
		shardId:     shardId,
		storage:     stoarge,
		stop:        make(chan struct{}),
		wg:          wg,
	}
	ex.tw = util.NewTickWorker("delay-executor-"+shardId, 1*time.Second, ex.stop, ex.handle, ex.wg)
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

func (ex *delayExecutor) handle() {
	actions, err := ex.storage.PollDelay()
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions {
		ex.flowService.ExecuteAction(action.WorkflowName, action.FlowId, "default", action.ActionId, nil)
	}
}
