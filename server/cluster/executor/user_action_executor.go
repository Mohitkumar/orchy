package executor

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ Executor = new(userActionExecutor)

type userActionExecutor struct {
	shard persistence.Shard
	wg    *sync.WaitGroup
	tw    *util.TickWorker
	stop  chan struct{}
}

func NewUserActionExecutor(shard persistence.Shard, wg *sync.WaitGroup) *userActionExecutor {
	ex := &userActionExecutor{
		shard: shard,
		stop:  make(chan struct{}),
		wg:    wg,
	}
	ex.tw = util.NewTickWorker(ex.Name(), 1*time.Second, ex.stop, ex.handle, ex.wg)
	return ex
}

func (ex *userActionExecutor) Name() string {
	return "user-action-executor-" + ex.shard.GetShardId()
}

func (ex *userActionExecutor) Start() {
	if ex.IsRunning() {
		return
	}
	ex.tw.Start()
}

func (ex *userActionExecutor) IsRunning() bool {
	return ex.tw.IsRunning()
}

func (ex *userActionExecutor) Stop() {
	if !ex.IsRunning() {
		return
	}
	ex.stop <- struct{}{}
}

func (ex *userActionExecutor) handle() {
	actions, err := ex.shard.PollAction("user", 100)
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions {
		logger.Info("running action", zap.String("name", action.ActionName), zap.String("workflow", action.WorkflowName), zap.String("id", action.FlowId))
	}
	ex.shard.GetExternalQueue().Push(actions)
}
