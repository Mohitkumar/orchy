package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	rd "github.com/go-redis/redis/v9"
	"github.com/mohitkumar/orchy/model"
	"github.com/mohitkumar/orchy/persistence"
	"github.com/mohitkumar/orchy/shard"
	"github.com/mohitkumar/orchy/util"
)

const WORKFLOW_KEY string = "FLOW"

var _ shard.Storage = new(redisShardStorage)

type redisShardStorage struct {
	shardId string
	*baseDao
	encoderDecoder util.EncoderDecoder[model.FlowContext]
}

func NewRedisStorage(baseDao *baseDao, encoderDecoder util.EncoderDecoder[model.FlowContext], shardId string) *redisShardStorage {
	return &redisShardStorage{
		baseDao:        baseDao,
		encoderDecoder: encoderDecoder,
		shardId:        shardId,
	}
}

func (r *redisShardStorage) SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error {
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

func (r *redisShardStorage) GetFlowContext(wfName string, flowId string) (*model.FlowContext, error) {
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

func (r *redisShardStorage) DeleteFlowContext(wfName string, flowId string) error {
	key := r.baseDao.getNamespaceKey(WORKFLOW_KEY, wfName, r.shardId)
	ctx := context.Background()
	err := r.baseDao.redisClient.HDel(ctx, key, flowId).Err()
	if err != nil {
		return persistence.StorageLayerError{Message: err.Error()}
	}
	return nil
}

func (r *redisShardStorage) SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, actions []model.ActionExecutionRequest) error {
	var messagesUser []rd.Z
	var messagesSystem []rd.Z
	for _, action := range actions {
		member := rd.Z{
			Score:  float64(time.Now().UnixMilli()),
			Member: fmt.Sprintf("%s:%s:%s:%d", wfName, flowId, action.ActionName, action.ActionId),
		}
		if action.ActionType == model.ACTION_TYPE_SYSTEM {
			messagesSystem = append(messagesSystem, member)
		} else {
			messagesUser = append(messagesUser, member)
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
			err = pipe.ZAdd(ctx, queueNameUser, messagesUser...).Err()
		}
		if len(messagesSystem) != 0 {
			err = pipe.ZAdd(ctx, queueNameSystem, messagesSystem...).Err()
		}
		return err
	})

	if err != nil {
		return persistence.StorageLayerError{Message: err.Error()}
	}
	return nil
}
func (r *redisShardStorage) PollAction(actionType string, batchSize int) ([]model.ActionExecutionRequest, error) {
	queueName := r.getNamespaceKey(actionType, r.shardId)
	ctx := context.Background()
	values, err := r.redisClient.ZRange(ctx, queueName, 0, int64(batchSize-1)).Result()
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

func (r *redisShardStorage) Ack(actionType string, actions []model.ActionExecutionRequest) error {
	queueName := r.getNamespaceKey(actionType, r.shardId)
	ctx := context.Background()
	var messages []string
	for _, act := range actions {
		message := fmt.Sprintf("%s:%s:%s:%d", act.WorkflowName, act.FlowId, act.ActionName, act.ActionId)
		messages = append(messages, message)
	}
	err := r.redisClient.ZRem(ctx, queueName, messages).Err()
	if err != nil {
		return persistence.StorageLayerError{Message: err.Error()}
	}
	return nil
}

func (r *redisShardStorage) Retry(wfName string, flowId string, actionName string, actionId int, delay time.Duration) error {
	message := fmt.Sprintf("%s:%s:%s:%d", wfName, flowId, actionName, actionId)
	queueName := r.getNamespaceKey("retry", r.shardId)
	return r.addToSortedSet(queueName, []string{message}, delay)
}
func (r *redisShardStorage) PollRetry() ([]model.ActionExecutionRequest, error) {
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
func (r *redisShardStorage) Delay(wfName string, flowId string, actionName string, actionId int, delay time.Duration) error {
	message := fmt.Sprintf("%s:%s:%s:%d", wfName, flowId, actionName, actionId)
	queueName := r.getNamespaceKey("delay", r.shardId)
	return r.addToSortedSet(queueName, []string{message}, delay)
}
func (r *redisShardStorage) PollDelay() ([]model.ActionExecutionRequest, error) {
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

func (r *redisShardStorage) Timeout(wfName string, flowId string, actionName string, actionId int, delay time.Duration) error {
	message := fmt.Sprintf("%s:%s:%s:%d", wfName, flowId, actionName, actionId)
	queueName := r.getNamespaceKey("timeout", r.shardId)
	return r.addToSortedSet(queueName, []string{message}, delay)
}

func (r *redisShardStorage) PollTimeout() ([]model.ActionExecutionRequest, error) {
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

func (r *redisShardStorage) addToSortedSet(key string, messages []string, delay time.Duration) error {
	ctx := context.Background()
	currentTime := time.Now().Add(delay).UnixMilli()
	var members []rd.Z
	for _, message := range messages {
		member := rd.Z{
			Score:  float64(currentTime),
			Member: message,
		}
		members = append(members, member)
	}

	err := r.redisClient.ZAdd(ctx, key, members...).Err()
	if err != nil {
		return persistence.StorageLayerError{Message: err.Error()}
	}
	return nil
}

func (r *redisShardStorage) getExpiredFromSortedSet(key string) ([]string, error) {
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
