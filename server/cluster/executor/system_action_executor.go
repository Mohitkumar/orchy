package executor

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ Executor = new(systemActionExecutor)

type systemActionExecutor struct {
	flowService *flow.FlowService
	shard       persistence.Shard
	tw          *util.TickWorker
	wg          *sync.WaitGroup
	stop        chan struct{}
}

func NewSystemActionExecutor(flowService *flow.FlowService, shard persistence.Shard, wg *sync.WaitGroup) *systemActionExecutor {
	ex := &systemActionExecutor{
		flowService: flowService,
		shard:       shard,
		stop:        make(chan struct{}),
		wg:          wg,
	}
	ex.tw = util.NewTickWorker(ex.Name(), 1*time.Second, ex.stop, ex.handle, ex.wg)
	return ex
}

func (ex *systemActionExecutor) Name() string {
	return "system-action-executor-" + ex.shard.GetShardId()
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
	actions, err := ex.shard.PollAction("system", 10)
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions {
		ex.flowService.ExecuteSystemAction(action.WorkflowName, action.FlowId, 1, action.ActionId)
	}
}
