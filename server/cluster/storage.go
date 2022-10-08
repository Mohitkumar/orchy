package cluster

import (
	"strconv"
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
)

type Queue interface {
	Push(queueName string, flowId string, mesage []byte) error
	Pop(queuName string, batchSize int) ([]string, error)
}

var _ Queue = new(clusterQueue)

type clusterQueue struct {
	redisQueue persistence.Queue
	ring       *Ring
	mu         sync.Mutex
}

func NewQueue(queue persistence.Queue, ring *Ring) *clusterQueue {
	return &clusterQueue{
		redisQueue: queue,
		ring:       ring,
	}
}

func (cq *clusterQueue) Push(queueName string, flowId string, mesage []byte) error {
	partition := strconv.Itoa(cq.ring.GetPartition(flowId))
	return cq.redisQueue.Push(queueName, partition, mesage)
}

func (cq *clusterQueue) Pop(queueName string, batchSize int) ([]string, error) {
	result := make([]string, 0)
	partitions := cq.ring.GetPartitions()
	for _, partition := range partitions {
		if len(result) < batchSize {
			numOfItemsToFetch := batchSize - len(result)
			items, err := cq.redisQueue.Pop(queueName, strconv.Itoa(partition), numOfItemsToFetch)
			if err != nil {
				return nil, err
			}
			result = append(result, items...)
		} else {
			break
		}
	}

	return result, nil
}

type DelayQueue interface {
	Push(queueName string, flowId string, mesage []byte) error
	Pop(queueName string) ([]string, error)
	PushWithDelay(queueName string, flowId string, delay time.Duration, message []byte) error
}

type clusterDelayQueue struct {
	redisQueue persistence.DelayQueue
	ring       *Ring
}

var _ DelayQueue = new(clusterDelayQueue)

func NewDelayQueue(queue persistence.DelayQueue, ring *Ring) *clusterDelayQueue {
	return &clusterDelayQueue{
		redisQueue: queue,
		ring:       ring,
	}
}
func (dq *clusterDelayQueue) Push(queueName string, flowId string, mesage []byte) error {
	partition := strconv.Itoa(dq.ring.GetPartition(flowId))
	return dq.redisQueue.Push(queueName, partition, mesage)
}

func (dq *clusterDelayQueue) Pop(queueName string) ([]string, error) {
	partitions := dq.ring.GetPartitions()
	result := make([]string, 0)
	for _, part := range partitions {
		res, err := dq.redisQueue.Pop(queueName, strconv.Itoa(part))
		if err != nil {
			return nil, err
		}
		result = append(result, res...)
	}
	return result, nil
}

func (dq *clusterDelayQueue) PushWithDelay(queueName string, flowId string, delay time.Duration, message []byte) error {
	partition := strconv.Itoa(dq.ring.GetPartition(flowId))
	return dq.redisQueue.PushWithDelay(queueName, partition, delay, message)
}

type FlowDao interface {
	SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error
	CreateAndSaveFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error)
	AddActionOutputToFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error)
	GetFlowContext(wfName string, flowId string) (*model.FlowContext, error)
	DeleteFlowContext(wfName string, flowId string) error
}

type clusterFlowDao struct {
	flowDao persistence.FlowDao
	ring    *Ring
}

var _ FlowDao = new(clusterFlowDao)

func NewFlowDao(dao persistence.FlowDao, ring *Ring) *clusterFlowDao {
	return &clusterFlowDao{
		flowDao: dao,
		ring:    ring,
	}
}

func (cd *clusterFlowDao) SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error {
	partition := strconv.Itoa(cd.ring.GetPartition(flowId))
	return cd.flowDao.SaveFlowContext(wfName, flowId, partition, flowCtx)
}

func (cd *clusterFlowDao) CreateAndSaveFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error) {
	partition := strconv.Itoa(cd.ring.GetPartition(flowId))
	return cd.flowDao.CreateAndSaveFlowContext(wFname, flowId, partition, action, dataMap)
}

func (cd *clusterFlowDao) AddActionOutputToFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error) {
	partition := strconv.Itoa(cd.ring.GetPartition(flowId))
	return cd.flowDao.AddActionOutputToFlowContext(wFname, flowId, partition, action, dataMap)
}
func (cd *clusterFlowDao) GetFlowContext(wfName string, flowId string) (*model.FlowContext, error) {
	partition := strconv.Itoa(cd.ring.GetPartition(flowId))
	return cd.flowDao.GetFlowContext(wfName, flowId, partition)
}
func (cd *clusterFlowDao) DeleteFlowContext(wfName string, flowId string) error {
	partition := strconv.Itoa(cd.ring.GetPartition(flowId))
	return cd.flowDao.DeleteFlowContext(wfName, flowId, partition)
}
