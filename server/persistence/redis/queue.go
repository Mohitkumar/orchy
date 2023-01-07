package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v9"
	rd "github.com/go-redis/redis/v9"
	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type redisQueue struct {
	*baseDao
}

func NewRedisQueue(config Config) *redisQueue {
	return &redisQueue{
		baseDao: newBaseDao(config),
	}
}
func (rq *redisQueue) Push(actions []model.ActionExecutionRequest) error {
	ctx := context.Background()
	groups := rq.groupByActionName(actions)
	_, err := rq.baseDao.redisClient.TxPipelined(ctx, func(pipe rd.Pipeliner) error {
		var err error
		for k, v := range groups {
			queueName := rq.getNamespaceKey(k)
			err = pipe.SAdd(ctx, queueName, v).Err()
		}
		return err
	})
	if err != nil {
		logger.Error("error while push to redis list", zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (rq *redisQueue) Poll(actionName string, batchSize int) ([]model.ActionExecutionRequest, error) {
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

func (rq *redisQueue) groupByActionName(actions []model.ActionExecutionRequest) map[string][]string {
	out := make(map[string][]string)
	for _, act := range actions {
		if _, ok := out[act.ActionName]; !ok {
			out[act.ActionName] = []string{fmt.Sprintf("%s:%s:%s:%d", act.WorkflowName, act.FlowId, act.ActionName, act.ActionId)}
		} else {
			list := out[act.ActionName]
			list = append(list, fmt.Sprintf("%s:%s:%s:%d", act.WorkflowName, act.FlowId, act.ActionName, act.ActionId))
		}
	}
	return out
}
