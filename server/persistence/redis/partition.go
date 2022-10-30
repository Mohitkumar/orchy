package redis

import (
	"strconv"

	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
)

func NewRedisPartitions(config Config, encoderDecoder util.EncoderDecoder[model.FlowContext], partitionCount int) *persistence.Partitions {
	parts := &persistence.Partitions{
		Partitions: make(map[int]persistence.Partition, partitionCount),
	}
	for i := 0; i < partitionCount; i++ {
		parts.Partitions[i] = NewRedisPartition(config, encoderDecoder, strconv.Itoa(i))
	}
	return parts
}

type redisPartition struct {
	flowDao     *redisFlowDao
	queue       *redisQueue
	delayQueue  *redisDelayQueue
	partitionId string
}

var _ persistence.Partition = new(redisPartition)

func NewRedisPartition(config Config, encoderDecoder util.EncoderDecoder[model.FlowContext], partId string) *redisPartition {
	baseDao := newBaseDao(config)
	return &redisPartition{
		flowDao:     NewRedisFlowDao(*baseDao, encoderDecoder, partId),
		queue:       NewRedisQueue(*baseDao, partId),
		delayQueue:  NewRedisDelayQueue(*baseDao, partId),
		partitionId: partId,
	}
}

func (p *redisPartition) GetFlowDao() persistence.FlowDao {
	return p.flowDao
}
func (p *redisPartition) GetQueue() persistence.Queue {
	return p.queue
}
func (p *redisPartition) GetDelayQueue() persistence.DelayQueue {
	return p.delayQueue
}
