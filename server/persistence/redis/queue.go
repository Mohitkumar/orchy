package redis

import (
	"context"
	"errors"
	"sync"

	"github.com/go-redis/redis/v9"
	"github.com/golang/protobuf/proto"
	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
	"go.uber.org/zap"
)

type redisQueue struct {
	*baseDao
	queueName string
	mu        sync.Mutex
}

func NewRedisQueue(config Config) *redisQueue {
	return &redisQueue{
		baseDao: newBaseDao(config),
	}
}
func (rq *redisQueue) Push(action *api.Action) error {
	queueName := rq.getNamespaceKey(action.ActionName)
	msg, err := proto.Marshal(action)
	if err != nil {
		return err
	}
	ctx := context.Background()

	err = rq.redisClient.LPush(ctx, queueName, msg).Err()
	if err != nil {
		logger.Error("error while push to redis list", zap.String("queue", queueName), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (rq *redisQueue) Poll(actionName string, batchSize int) (*api.Actions, error) {
	queueName := rq.getNamespaceKey(actionName)
	ctx := context.Background()
	var out []*api.Action
	values, err := rq.redisClient.LPopCount(ctx, queueName, batchSize).Result()
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
