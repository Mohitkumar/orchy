package redis

import (
	"context"
	"errors"
	"sync"

	"github.com/go-redis/redis/v9"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
	"go.uber.org/zap"
)

type redisQueue struct {
	baseDao
	mu         sync.Mutex
	partitonId string
}

var _ persistence.Queue = new(redisQueue)

func NewRedisQueue(baseDao baseDao, partitionId string) *redisQueue {
	return &redisQueue{
		baseDao:    baseDao,
		partitonId: partitionId,
	}
}
func (rq *redisQueue) Push(queueName string, mesage []byte) error {
	queueName = rq.getNamespaceKey(queueName, rq.partitonId)
	ctx := context.Background()

	err := rq.redisClient.LPush(ctx, queueName, mesage).Err()
	if err != nil {
		logger.Error("error while push to redis list", zap.String("queue", queueName), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (rq *redisQueue) Pop(queueName string, batchSize int) ([]string, error) {
	queueName = rq.getNamespaceKey(queueName, rq.partitonId)
	ctx := context.Background()
	res, err := rq.redisClient.LPopCount(ctx, queueName, batchSize).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []string{}, nil
		}
		logger.Error("error while pop from redis list", zap.String("queue", queueName), zap.Error(err))
		return nil, persistence.StorageLayerError{}
	}
	return res, nil
}
