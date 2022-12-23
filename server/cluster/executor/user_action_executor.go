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
	stop        chan struct{}
}

func NewUserActionExecutor(diContainer *container.DIContiner, shard persistence.Shard, wg *sync.WaitGroup) *userActionExecutor {
	return &userActionExecutor{
		diContainer: diContainer,
		shard:       shard,
		stop:        make(chan struct{}),
		wg:          wg,
	}
}

func (ex *userActionExecutor) Name() string {
	return "user-action-executor-" + ex.shard.GetShardId()
}

func (ex *userActionExecutor) Start() {
	tw := util.NewTickWorker(ex.Name(), 1*time.Second, ex.stop, ex.handle, ex.wg)
	tw.Start()
}

func (ex *userActionExecutor) Stop() {
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
		ex.diContainer.GetExternalQueue().Push(action)
		ex.diContainer.GetClusterStorage().Timeout(action, time.Duration(actionDef.TimeoutSeconds)*time.Second)
	}
}
