package redis

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v9"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
	"go.uber.org/zap"
)

type redisQueue struct {
	baseDao
}

var _ persistence.Queue = new(redisQueue)

func NewRedisQueue(config Config) *redisQueue {
	return &redisQueue{
		baseDao: *newBaseDao(config),
	}
}
func (rq *redisQueue) Push(queueName string, mesage []byte) error {
	queueName = rq.getNamespaceKey(queueName)
	ctx := context.Background()

	err := rq.redisClient.LPush(ctx, queueName, mesage).Err()
	if err != nil {
		logger.Error("error while push to redis list", zap.String("queue", queueName), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (rq *redisQueue) Pop(queueName string) ([]byte, error) {
	queueName = rq.getNamespaceKey(queueName)
	ctx := context.Background()
	res, err := rq.redisClient.LPop(ctx, queueName).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, persistence.EmptyQueueError{QueueName: queueName}
		}
		logger.Error("error while pop from redis list", zap.String("queue", queueName), zap.Error(err))

		return nil, persistence.StorageLayerError{}
	}
	return []byte(res), nil
}
