package redis

import (
	"context"
	"strconv"
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

const WORKFLOW_KEY string = "FLOW"

var _ persistence.Shard = new(redisShard)

type redisShard struct {
	shardId string
	*baseDao
	encoderDecoder util.EncoderDecoder[model.FlowContext]
}

func InitRedisShards(config Config, encoderDecoder util.EncoderDecoder[model.FlowContext], partitionCount int) *persistence.Shards {
	shards := &persistence.Shards{
		Shards: make(map[int]persistence.Shard, partitionCount),
	}
	for i := 0; i < partitionCount; i++ {
		shards.Shards[i] = NewRedisShard(config, encoderDecoder, strconv.Itoa(i))
	}
	return shards
}

func NewRedisShard(conf Config, encoderDecoder util.EncoderDecoder[model.FlowContext], shardId string) *redisShard {
	return &redisShard{
		baseDao:        newBaseDao(conf),
		encoderDecoder: encoderDecoder,
		shardId:        shardId,
	}
}

func (r *redisShard) CreateAndSaveFlowContext(wFname string, flowId string, action int, input map[string]any) (*model.FlowContext, error) {
	dataMap := make(map[string]any)
	dataMap["input"] = input
	flowCtx := &model.FlowContext{
		Id:            flowId,
		State:         model.RUNNING,
		CurrentAction: action,
		Data:          dataMap,
	}
	if err := r.SaveFlowContext(wFname, flowId, flowCtx); err != nil {
		return nil, err
	}

	return flowCtx, nil
}

func (r *redisShard) SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error {
	key := r.baseDao.getNamespaceKey(WORKFLOW_KEY, wfName, r.shardId)
	ctx := context.Background()
	data, err := r.encoderDecoder.Encode(*flowCtx)
	if err != nil {
		return err
	}
	if err := r.baseDao.redisClient.HSet(ctx, key, []string{flowId, string(data)}).Err(); err != nil {
		logger.Error("error in saving flow context", zap.String("flowName", wfName), zap.String("flowId", flowId), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (r *redisShard) GetFlowContext(wfName string, flowId string) (*model.FlowContext, error) {
	key := r.baseDao.getNamespaceKey(WORKFLOW_KEY, wfName, r.shardId)
	ctx := context.Background()
	flowCtxStr, err := r.baseDao.redisClient.HGet(ctx, key, flowId).Result()
	if err != nil {
		logger.Error("error in getting flow context", zap.String("flowName", wfName), zap.String("flowId", flowId), zap.Error(err))
		return nil, persistence.StorageLayerError{}
	}

	flowCtx, err := r.encoderDecoder.Decode([]byte(flowCtxStr))
	if err != nil {
		return nil, err
	}
	return flowCtx, nil
}

func (r *redisShard) DeleteFlowContext(wfName string, flowId string) error {
	key := r.baseDao.getNamespaceKey(WORKFLOW_KEY, wfName, r.shardId)
	ctx := context.Background()
	err := r.baseDao.redisClient.HDel(ctx, key, flowId).Err()
	if err != nil {
		logger.Error("error in deleting flow context", zap.String("flowName", wfName), zap.String("flowId", flowId), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (r *redisShard) DispatchAction(wfName string, flowId string, action *api.Action) error {

}
func (r *redisShard) SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, action *api.Action) error {

}
func (r *redisShard) PollAction(wfName string, flowId string, actionName string) (*api.Actions, error) {

}

func (r *redisShard) Retry(req *model.ActionExecutionRequest, delay time.Duration) error {

}
func (r *redisShard) PollRetry(batch int) (*model.ActionExecutionRequest, error) {

}
func (r *redisShard) Delay(req *model.ActionExecutionRequest, delay time.Duration) error {

}
func (r *redisShard) PollDelay(batch int) (*model.ActionExecutionRequest, error) {

}
func (r *redisShard) Timeout(req *model.ActionExecutionRequest, delay time.Duration) error {

}
func (r *redisShard) PollTimeout(batch int) (*model.ActionExecutionRequest, error) {

}
