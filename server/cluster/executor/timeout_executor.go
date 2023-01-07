package executor

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ Executor = new(systemActionExecutor)

type timeoutExecutor struct {
	flowService *flow.FlowService
	diContainer *container.DIContiner
	shard       persistence.Shard
	wg          *sync.WaitGroup
	tw          *util.TickWorker
	stop        chan struct{}
}

func NewTimeoutExecutor(flowService *flow.FlowService, diContainer *container.DIContiner, shard persistence.Shard, wg *sync.WaitGroup) *timeoutExecutor {
	ex := &timeoutExecutor{
		flowService: flowService,
		diContainer: diContainer,
		shard:       shard,
		stop:        make(chan struct{}),
		wg:          wg,
	}
	ex.tw = util.NewTickWorker(ex.Name(), 1*time.Second, ex.stop, ex.handle, ex.wg)
	return ex
}

func (ex *timeoutExecutor) Name() string {
	return "timeout-executor-" + ex.shard.GetShardId()
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
	actions, err := ex.shard.PollTimeout()
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions {
		parts := strings.Split(action, ":")
		actionId, _ := strconv.Atoi(parts[3])
		ex.flowService.RetryAction(parts[0], parts[1], actionId, "timeout")
	}
}
