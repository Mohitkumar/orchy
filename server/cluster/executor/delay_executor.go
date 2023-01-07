package executor

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ Executor = new(systemActionExecutor)

type delayExecutor struct {
	flowService *flow.FlowService
	shard       persistence.Shard
	wg          *sync.WaitGroup
	tw          *util.TickWorker
	stop        chan struct{}
}

func NewDelayExecutor(flowService *flow.FlowService, shard persistence.Shard, wg *sync.WaitGroup) *delayExecutor {
	ex := &delayExecutor{
		flowService: flowService,
		shard:       shard,
		stop:        make(chan struct{}),
		wg:          wg,
	}
	ex.tw = util.NewTickWorker(ex.Name(), 1*time.Second, ex.stop, ex.handle, ex.wg)
	return ex
}

func (ex *delayExecutor) Name() string {
	return "delay-executor-" + ex.shard.GetShardId()
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
	actions, err := ex.shard.PollDelay()
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions {
		parts := strings.Split(action, ":")
		actionId, _ := strconv.Atoi(parts[3])
		ex.flowService.ExecuteAction(parts[0], parts[1], "default", actionId, nil)
	}
}
