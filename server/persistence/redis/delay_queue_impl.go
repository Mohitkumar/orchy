package redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis/v9"
	rd "github.com/go-redis/redis/v9"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
	"go.uber.org/zap"
)

type redisDelayQueue struct {
	baseDao
}

var _ persistence.DelayQueue = new(redisDelayQueue)

func NewRedisDelayQueue(config Config) *redisDelayQueue {
	return &redisDelayQueue{
		baseDao: *newBaseDao(config),
	}
}

func (rq *redisDelayQueue) Push(queueName string, message []byte) error {
	queueName = rq.getNamespaceKey(queueName)
	ctx := context.Background()
	currentTime := time.Now().UnixMilli()
	member := rd.Z{
		Score:  float64(currentTime),
		Member: message,
	}
	err := rq.redisClient.ZAdd(ctx, queueName, member).Err()
	if err != nil {
		logger.Error("error while push to redis list", zap.String("queue", queueName), zap.Error(err))
		return persistence.StorageLayerError{Message: err.Error()}
	}
	return nil
}

func (rq *redisDelayQueue) PushWithDelay(queueName string, delay time.Duration, message []byte) error {
	queueName = rq.getNamespaceKey(queueName)
	ctx := context.Background()
	currentTime := time.Now().Add(delay).UnixMilli()
	member := rd.Z{
		Score:  float64(currentTime),
		Member: message,
	}
	err := rq.redisClient.ZAdd(ctx, queueName, member).Err()
	if err != nil {
		logger.Error("error while push to redis list", zap.String("queue", queueName), zap.Error(err))
		return persistence.StorageLayerError{Message: err.Error()}
	}
	return nil
}

func (rq *redisDelayQueue) Pop(queueName string) ([]string, error) {
	queueName = rq.getNamespaceKey(queueName)
	ctx := context.Background()
	currentTime := time.Now().UnixMilli()
	pipe := rq.redisClient.Pipeline()

	opt := &rd.ZRangeBy{
		Min: strconv.Itoa(0),
		Max: strconv.FormatInt(currentTime, 10),
	}
	zr := pipe.ZRangeByScore(ctx, queueName, opt)
	pipe.ZRemRangeByScore(ctx, queueName, strconv.Itoa(0), strconv.FormatInt(currentTime, 10))

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Error("error while pop from redis list", zap.String("queue", queueName), zap.Error(err))

		return nil, persistence.StorageLayerError{Message: err.Error()}
	}

	res, err := zr.Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, persistence.EmptyQueueError{QueueName: queueName}
		}
		logger.Error("error while pop from redis list", zap.String("queue", queueName), zap.Error(err))

		return nil, persistence.StorageLayerError{Message: err.Error()}
	}
	if len(res) == 0 {
		return nil, persistence.EmptyQueueError{QueueName: queueName}
	}
	return res, nil
}
