package executor

import (
	"sync"
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/metadata"
	"github.com/mohitkumar/orchy/shard"
	"github.com/mohitkumar/orchy/util"
	"go.uber.org/zap"
)

var _ shard.Executor = new(userActionExecutor)

type userActionExecutor struct {
	shardId         string
	storage         shard.Storage
	externalQueue   shard.ExternalQueue
	metadataService metadata.MetadataService
	wg              *sync.WaitGroup
	tw              *util.TickWorker
	stop            chan struct{}
	batchSize       int
}

func NewUserActionExecutor(shardId string, storage shard.Storage, metadataService metadata.MetadataService, externalQueue shard.ExternalQueue, batchSize int, wg *sync.WaitGroup) *userActionExecutor {
	ex := &userActionExecutor{
		shardId:         shardId,
		storage:         storage,
		externalQueue:   externalQueue,
		metadataService: metadataService,
		stop:            make(chan struct{}),
		batchSize:       batchSize,
		wg:              wg,
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
	actions, err := ex.storage.PollAction("user", ex.batchSize)
	if err != nil {
		logger.Error("error while polling user actions", zap.Error(err))
	}
	for _, action := range actions {
		flowCtx, err := ex.storage.GetFlowContext(action.WorkflowName, action.FlowId)
		if err != nil {
			continue
		}
		logger.Info("running action", zap.String("name", action.ActionName), zap.String("workflow", action.WorkflowName), zap.String("id", action.FlowId))
		flow, err := ex.metadataService.GetFlow(action.WorkflowName, action.FlowId)
		if err != nil {
			continue
		}
		currentAction := flow.Actions[action.ActionId]
		act := &api.Action{
			WorkflowName: action.WorkflowName,
			FlowId:       action.FlowId,
			Data:         util.ConvertToProto(util.ResolveInputParams(flowCtx, currentAction.GetInputParams())),
			ActionId:     int32(action.ActionId),
			ActionName:   action.ActionName,
		}
		ex.externalQueue.Push(act)
	}
	if len(actions) != 0 {
		err = ex.storage.Ack("user", actions)
		if err != nil {
			logger.Error("error while ack user actions", zap.Error(err))
		}
	}
}
