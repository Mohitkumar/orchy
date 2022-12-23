package redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis/v9"
	rd "github.com/go-redis/redis/v9"
	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
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

func (r *redisShard) GetShardId() string {
	return r.shardId
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

func (r *redisShard) DispatchAction(action *api.Action, actionType string) error {
	message, err := proto.Marshal(action)
	if err != nil {
		return err
	}
	queueName := r.getNamespaceKey(actionType, r.shardId)
	ctx := context.Background()
	err = r.baseDao.redisClient.LPush(ctx, queueName, message).Err()
	if err != nil {
		logger.Error("error while push to redis list", zap.String("queue", queueName), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (r *redisShard) SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, action *api.Action, actionType string) error {
	message, err := proto.Marshal(action)
	if err != nil {
		return err
	}
	key := r.baseDao.getNamespaceKey(WORKFLOW_KEY, wfName, r.shardId)
	queueName := r.getNamespaceKey(actionType, r.shardId)
	ctx := context.Background()
	data, err := r.encoderDecoder.Encode(*flowCtx)
	if err != nil {
		return err
	}
	_, err = r.baseDao.redisClient.TxPipelined(ctx, func(pipe rd.Pipeliner) error {
		err := pipe.HSet(ctx, key, []string{flowId, string(data)}).Err()
		err = pipe.LPush(ctx, queueName, message).Err()
		return err
	})

	if err != nil {
		logger.Error("error in saving flow context", zap.String("flowName", wfName), zap.String("flowId", flowId), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}
func (r *redisShard) PollAction(actionType string, batchSize int) (*api.Actions, error) {
	queueName := r.getNamespaceKey(actionType, r.shardId)
	ctx := context.Background()
	var out []*api.Action
	values, err := r.redisClient.LPopCount(ctx, queueName, batchSize).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return &api.Actions{Actions: out}, nil
		}
		logger.Error("error while pop from redis list", zap.String("queue", queueName), zap.Error(err))
		return nil, persistence.StorageLayerError{}
	}
	for _, value := range values {
		action := &api.Action{}
		proto.Unmarshal([]byte(value), action)
		out = append(out, action)
	}
	return &api.Actions{Actions: out}, nil
}

func (r *redisShard) Retry(action *api.Action, delay time.Duration) error {
	message, err := proto.Marshal(action)
	if err != nil {
		return err
	}
	queueName := r.getNamespaceKey("retry", r.shardId)
	return r.addToSortedSet(queueName, message, delay)
}
func (r *redisShard) PollRetry(batch int) (*api.Actions, error) {
	queueName := r.getNamespaceKey("retry", r.shardId)
	values, err := r.getExpiredFromSortedSet(queueName, batch)
	if err != nil {
		return nil, err
	}
	var out []*api.Action
	for _, value := range values {
		action := &api.Action{}
		proto.Unmarshal([]byte(value), action)
		out = append(out, action)
	}
	return &api.Actions{Actions: out}, nil
}
func (r *redisShard) Delay(action *api.Action, delay time.Duration) error {
	message, err := proto.Marshal(action)
	if err != nil {
		return err
	}
	queueName := r.getNamespaceKey("delay", r.shardId)
	return r.addToSortedSet(queueName, message, delay)
}
func (r *redisShard) PollDelay(batch int) (*api.Actions, error) {
	queueName := r.getNamespaceKey("delay", r.shardId)
	values, err := r.getExpiredFromSortedSet(queueName, batch)
	if err != nil {
		return nil, err
	}
	var out []*api.Action
	for _, value := range values {
		action := &api.Action{}
		proto.Unmarshal([]byte(value), action)
		out = append(out, action)
	}
	return &api.Actions{Actions: out}, nil
}

func (r *redisShard) Timeout(action *api.Action, delay time.Duration) error {
	message, err := proto.Marshal(action)
	if err != nil {
		return err
	}
	queueName := r.getNamespaceKey("timeout", r.shardId)
	return r.addToSortedSet(queueName, message, delay)
}

func (r *redisShard) PollTimeout(batch int) (*api.Actions, error) {
	queueName := r.getNamespaceKey("timeout", r.shardId)
	values, err := r.getExpiredFromSortedSet(queueName, batch)
	if err != nil {
		return nil, err
	}
	var out []*api.Action
	for _, value := range values {
		action := &api.Action{}
		proto.Unmarshal([]byte(value), action)
		out = append(out, action)
	}
	return &api.Actions{Actions: out}, nil
}

func (r *redisShard) addToSortedSet(key string, message []byte, delay time.Duration) error {
	ctx := context.Background()
	currentTime := time.Now().Add(delay).UnixMilli()
	member := rd.Z{
		Score:  float64(currentTime),
		Member: message,
	}
	err := r.redisClient.ZAdd(ctx, key, member).Err()
	if err != nil {
		logger.Error("error while add to sorted set", zap.String("key", key), zap.Error(err))
		return persistence.StorageLayerError{Message: err.Error()}
	}
	return nil
}

func (r *redisShard) getExpiredFromSortedSet(key string, batch int) ([]string, error) {
	ctx := context.Background()
	currentTime := time.Now().UnixMilli()
	opt := &rd.ZRangeBy{
		Min: strconv.Itoa(0),
		Max: strconv.FormatInt(currentTime, 10),
	}
	var result []string
	_, err := r.baseDao.redisClient.TxPipelined(ctx, func(pipe rd.Pipeliner) error {
		res, err := pipe.ZRangeByScore(ctx, key, opt).Result()
		result = append(result, res...)
		err = pipe.ZRemRangeByScore(ctx, key, strconv.Itoa(0), strconv.FormatInt(currentTime, 10)).Err()
		return err
	})

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []string{}, nil
		}
		logger.Error("error while getting from redis sorted set", zap.String("key", key), zap.Error(err))
		return nil, persistence.StorageLayerError{Message: err.Error()}
	}
	return result, nil
}
