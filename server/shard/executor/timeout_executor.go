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

type timeoutExecutor struct {
	flowService *flow.FlowService
	shardId     string
	storage     shard.Storage
	wg          *sync.WaitGroup
	tw          *util.TickWorker
	stop        chan struct{}
}

func NewTimeoutExecutor(shardId string, storage shard.Storage, flowService *flow.FlowService, wg *sync.WaitGroup) *timeoutExecutor {
	ex := &timeoutExecutor{
		flowService: flowService,
		shardId:     shardId,
		storage:     storage,
		stop:        make(chan struct{}),
		wg:          wg,
	}
	ex.tw = util.NewTickWorker("timeout-executor-"+shardId, 1*time.Second, ex.stop, ex.handle, ex.wg)
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

func (ex *timeoutExecutor) handle() {
	actions, err := ex.storage.PollTimeout()
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions {
		ex.flowService.RetryAction(action.WorkflowName, action.FlowId, action.ActionId, "timeout")
	}
}
