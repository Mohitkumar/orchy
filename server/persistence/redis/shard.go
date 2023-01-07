package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	rd "github.com/go-redis/redis/v9"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
)

const WORKFLOW_KEY string = "FLOW"

var _ persistence.Shard = new(redisShard)

type redisShard struct {
	shardId string
	*baseDao
	encoderDecoder util.EncoderDecoder[model.FlowContext]
	externalQueue  persistence.ExternalQueue
}

func InitRedisShards(config Config, externalQueue persistence.ExternalQueue, encoderDecoder util.EncoderDecoder[model.FlowContext], partitionCount int) *persistence.Shards {
	shards := &persistence.Shards{
		Shards: make(map[int]persistence.Shard, partitionCount),
	}
	for i := 0; i < partitionCount; i++ {
		shards.Shards[i] = NewRedisShard(config, externalQueue, encoderDecoder, strconv.Itoa(i))
	}
	return shards
}

func NewRedisShard(conf Config, externalQueue persistence.ExternalQueue, encoderDecoder util.EncoderDecoder[model.FlowContext], shardId string) *redisShard {
	return &redisShard{
		baseDao:        newBaseDao(conf),
		encoderDecoder: encoderDecoder,
		shardId:        shardId,
		externalQueue:  externalQueue,
	}
}

func (r *redisShard) GetShardId() string {
	return r.shardId
}

func (r *redisShard) SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error {
	key := r.baseDao.getNamespaceKey(WORKFLOW_KEY, wfName, r.shardId)
	ctx := context.Background()
	data, err := r.encoderDecoder.Encode(*flowCtx)
	if err != nil {
		return err
	}
	if err := r.baseDao.redisClient.HSet(ctx, key, []string{flowId, string(data)}).Err(); err != nil {
		return persistence.StorageLayerError{Message: err.Error()}
	}
	return nil
}

func (r *redisShard) GetFlowContext(wfName string, flowId string) (*model.FlowContext, error) {
	key := r.baseDao.getNamespaceKey(WORKFLOW_KEY, wfName, r.shardId)
	ctx := context.Background()
	flowCtxStr, err := r.baseDao.redisClient.HGet(ctx, key, flowId).Result()
	if err != nil {
		return nil, persistence.StorageLayerError{Message: err.Error()}
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
		return persistence.StorageLayerError{Message: err.Error()}
	}
	return nil
}

func (r *redisShard) SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, actions []model.ActionExecutionRequest) error {
	var messagesUser []string
	var messagesSystem []string
	for _, action := range actions {
		message := fmt.Sprintf("%s:%s:%s:%d", wfName, flowId, action.ActionName, action.ActionId)
		if action.ActionType == model.ACTION_TYPE_SYSTEM {
			messagesSystem = append(messagesSystem, string(message))
		} else {
			messagesUser = append(messagesUser, string(message))
		}
	}
	key := r.baseDao.getNamespaceKey(WORKFLOW_KEY, wfName, r.shardId)
	queueNameSystem := r.getNamespaceKey("system", r.shardId)
	queueNameUser := r.getNamespaceKey("user", r.shardId)
	ctx := context.Background()
	data, _ := r.encoderDecoder.Encode(*flowCtx)
	_, err := r.baseDao.redisClient.TxPipelined(ctx, func(pipe rd.Pipeliner) error {
		err := pipe.HSet(ctx, key, []string{flowId, string(data)}).Err()
		if len(messagesUser) != 0 {
			err = pipe.LPush(ctx, queueNameUser, messagesUser).Err()
		}
		if len(messagesSystem) != 0 {
			err = pipe.LPush(ctx, queueNameSystem, messagesSystem).Err()
		}
		return err
	})

	if err != nil {
		return persistence.StorageLayerError{Message: err.Error()}
	}
	return nil
}
func (r *redisShard) PollAction(actionType string, batchSize int) ([]model.ActionExecutionRequest, error) {
	queueName := r.getNamespaceKey(actionType, r.shardId)
	ctx := context.Background()
	values, err := r.redisClient.LPopCount(ctx, queueName, batchSize).Result()
	if err != nil {
		if errors.Is(err, rd.Nil) {
			return []model.ActionExecutionRequest{}, nil
		}
		return nil, persistence.StorageLayerError{Message: err.Error()}
	}
	var resList []model.ActionExecutionRequest
	for _, v := range values {
		parts := strings.Split(v, ":")
		actionId, _ := strconv.Atoi(parts[3])
		res := model.ActionExecutionRequest{
			WorkflowName: parts[0],
			FlowId:       parts[1],
			ActionName:   parts[2],
			ActionId:     actionId,
		}
		resList = append(resList, res)
	}
	return resList, nil
}

func (r *redisShard) Retry(wfName string, flowId string, actionId int, delay time.Duration) error {
	message := fmt.Sprintf("%s:%s:%d", wfName, flowId, actionId)
	queueName := r.getNamespaceKey("retry", r.shardId)
	return r.addToSortedSet(queueName, []byte(message), delay)
}
func (r *redisShard) PollRetry() ([]model.ActionExecutionRequest, error) {
	queueName := r.getNamespaceKey("retry", r.shardId)
	values, err := r.getExpiredFromSortedSet(queueName)
	if err != nil {
		return nil, err
	}
	var resList []model.ActionExecutionRequest
	for _, v := range values {
		parts := strings.Split(v, ":")
		actionId, _ := strconv.Atoi(parts[3])
		res := model.ActionExecutionRequest{
			WorkflowName: parts[0],
			FlowId:       parts[1],
			ActionName:   parts[2],
			ActionId:     actionId,
		}
		resList = append(resList, res)
	}
	return resList, nil
}
func (r *redisShard) Delay(wfName string, flowId string, actionId int, delay time.Duration) error {
	message := fmt.Sprintf("%s:%s:%d", wfName, flowId, actionId)
	queueName := r.getNamespaceKey("delay", r.shardId)
	return r.addToSortedSet(queueName, []byte(message), delay)
}
func (r *redisShard) PollDelay() ([]model.ActionExecutionRequest, error) {
	queueName := r.getNamespaceKey("delay", r.shardId)
	values, err := r.getExpiredFromSortedSet(queueName)
	if err != nil {
		return nil, err
	}
	var resList []model.ActionExecutionRequest
	for _, v := range values {
		parts := strings.Split(v, ":")
		actionId, _ := strconv.Atoi(parts[3])
		res := model.ActionExecutionRequest{
			WorkflowName: parts[0],
			FlowId:       parts[1],
			ActionName:   parts[2],
			ActionId:     actionId,
		}
		resList = append(resList, res)
	}
	return resList, nil
}

func (r *redisShard) Timeout(wfName string, flowId string, actionId int, delay time.Duration) error {
	message := fmt.Sprintf("%s:%s:%d", wfName, flowId, actionId)
	queueName := r.getNamespaceKey("timeout", r.shardId)
	return r.addToSortedSet(queueName, []byte(message), delay)
}

func (r *redisShard) PollTimeout() ([]model.ActionExecutionRequest, error) {
	queueName := r.getNamespaceKey("timeout", r.shardId)
	values, err := r.getExpiredFromSortedSet(queueName)
	if err != nil {
		return nil, err
	}
	var resList []model.ActionExecutionRequest
	for _, v := range values {
		parts := strings.Split(v, ":")
		actionId, _ := strconv.Atoi(parts[3])
		res := model.ActionExecutionRequest{
			WorkflowName: parts[0],
			FlowId:       parts[1],
			ActionName:   parts[2],
			ActionId:     actionId,
		}
		resList = append(resList, res)
	}
	return resList, err
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
		return persistence.StorageLayerError{Message: err.Error()}
	}
	return nil
}

func (r *redisShard) getExpiredFromSortedSet(key string) ([]string, error) {
	ctx := context.Background()
	currentTime := time.Now().UnixMilli()
	opt := &rd.ZRangeBy{
		Min: strconv.Itoa(0),
		Max: strconv.FormatInt(currentTime, 10),
	}
	pipe := r.redisClient.Pipeline()
	zr := pipe.ZRangeByScore(ctx, key, opt)
	pipe.ZRemRangeByScore(ctx, key, strconv.Itoa(0), strconv.FormatInt(currentTime, 10))
	_, err := pipe.Exec(ctx)

	if err != nil {
		return nil, persistence.StorageLayerError{Message: err.Error()}
	}
	res, err := zr.Result()
	if err != nil {
		if errors.Is(err, rd.Nil) {
			return []string{}, nil
		}
		return nil, persistence.StorageLayerError{Message: err.Error()}
	}
	return res, nil
}

func (r *redisShard) GetExternalQueue() persistence.ExternalQueue {
	return r.externalQueue
}
