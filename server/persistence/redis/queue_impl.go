package redis

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/go-redis/redis/v9"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
	"go.uber.org/zap"
)

type redisQueue struct {
	baseDao
	mu               sync.Mutex
	currentPartition uint64
}

var _ persistence.Queue = new(redisQueue)

func NewRedisQueue(config Config) *redisQueue {
	return &redisQueue{
		baseDao: *newBaseDao(config),
	}
}
func (rq *redisQueue) Push(queueName string, flowId string, mesage []byte) error {
	partition := strconv.Itoa(rq.baseDao.getPartition(flowId))
	queueName = rq.getNamespaceKey(queueName, partition)
	ctx := context.Background()

	err := rq.redisClient.LPush(ctx, queueName, mesage).Err()
	if err != nil {
		logger.Error("error while push to redis list", zap.String("queue", queueName), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (rq *redisQueue) Pop(queueName string, batchSize int) ([]string, error) {

	result := make([]string, 0)
	for len(result) < batchSize {
		partition := rq.getNextPartition()
		queueName = rq.getNamespaceKey(queueName, strconv.Itoa(partition))
		numOfItemsToFetch := batchSize - len(result)
		items, err := rq.pop(queueName, numOfItemsToFetch)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}
	return result, nil
}

func (rq *redisQueue) pop(queueName string, batchSize int) ([]string, error) {
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

func (rq *redisQueue) getNextPartition() int {
	rq.mu.Lock()
	defer rq.mu.Unlock()
	partitions := rq.baseDao.ring.GetPartitions(rq.membership.GetLocalMemebr())
	partitions = sort.IntSlice(partitions)
	idx := sort.Search(len(partitions), func(i int) bool {
		return partitions[i] > int(rq.currentPartition)
	})
	if idx >= len(partitions) {
		idx = 0
	}
	nextPartition := partitions[idx]
	atomic.StoreUint64(&rq.currentPartition, uint64(nextPartition))
	return nextPartition
}
