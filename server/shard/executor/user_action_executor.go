package executor

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/shard"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ shard.Executor = new(userActionExecutor)

type userActionExecutor struct {
	shardId       string
	storage       shard.Storage
	externalQueue shard.ExternalQueue
	wg            *sync.WaitGroup
	tw            *util.TickWorker
	stop          chan struct{}
}

func NewUserActionExecutor(shardId string, storage shard.Storage, externalQueue shard.ExternalQueue, wg *sync.WaitGroup) *userActionExecutor {
	ex := &userActionExecutor{
		shardId:       shardId,
		storage:       storage,
		externalQueue: externalQueue,
		stop:          make(chan struct{}),
		wg:            wg,
	}
	ex.tw = util.NewTickWorker("user-action-executor-"+shardId, 1*time.Second, ex.stop, ex.handle, ex.wg)
	return ex
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
	actions, err := ex.storage.PollAction("user", 100)
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions {
		logger.Info("running action", zap.String("name", action.ActionName), zap.String("workflow", action.WorkflowName), zap.String("id", action.FlowId))
	}
	ex.externalQueue.Push(actions)
}
