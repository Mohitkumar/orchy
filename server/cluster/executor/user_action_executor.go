package executor

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ Executor = new(userActionExecutor)

type userActionExecutor struct {
	diContainer *container.DIContiner
	shard       persistence.Shard
	wg          *sync.WaitGroup
	tw          *util.TickWorker
	stop        chan struct{}
}

func NewUserActionExecutor(diContainer *container.DIContiner, shard persistence.Shard, wg *sync.WaitGroup) *userActionExecutor {
	ex := &userActionExecutor{
		diContainer: diContainer,
		shard:       shard,
		stop:        make(chan struct{}),
		wg:          wg,
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
	actions, err := ex.shard.PollAction("user", 10)
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions.Actions {
		actionDef, err := ex.diContainer.GetMetadataStorage().GetActionDefinition(action.ActionName)
		if err != nil {
			logger.Error("error getting action definition", zap.String("action", action.ActionName))
			continue
		}
		logger.Info("running action", zap.String("name", action.ActionName), zap.String("workflow", action.WorkflowName), zap.String("id", action.FlowId))
		ex.diContainer.GetExternalQueue().Push(action)
		ex.diContainer.GetClusterStorage().Timeout(action, time.Duration(actionDef.TimeoutSeconds)*time.Second)
	}
}
