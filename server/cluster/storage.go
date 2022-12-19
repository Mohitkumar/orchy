package cluster

import (
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
	parts *persistence.Partitions
	ring  *Ring
	mu    sync.Mutex
}

func NewQueue(parts *persistence.Partitions, ring *Ring) *clusterQueue {
	return &clusterQueue{
		parts: parts,
		ring:  ring,
	}
}

func (cq *clusterQueue) Push(queueName string, flowId string, mesage []byte) error {
	queue := cq.parts.GetPartition(cq.ring.GetPartition(flowId)).GetQueue()
	return queue.Push(queueName, mesage)
}

func (cq *clusterQueue) Pop(queueName string, batchSize int) ([]string, error) {
	result := make([]string, 0)
	partitions := cq.ring.GetPartitions()
	for _, partition := range partitions {
		if len(result) < batchSize {
			numOfItemsToFetch := batchSize - len(result)
			queue := cq.parts.GetPartition(partition).GetQueue()
			items, err := queue.Pop(queueName, numOfItemsToFetch)
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
	parts *persistence.Partitions
	ring  *Ring
}

var _ DelayQueue = new(clusterDelayQueue)

func NewDelayQueue(parts *persistence.Partitions, ring *Ring) *clusterDelayQueue {
	return &clusterDelayQueue{
		parts: parts,
		ring:  ring,
	}
}
func (dq *clusterDelayQueue) Push(queueName string, flowId string, mesage []byte) error {
	queue := dq.parts.GetPartition(dq.ring.GetPartition(flowId)).GetDelayQueue()
	return queue.Push(queueName, mesage)
}

func (dq *clusterDelayQueue) Pop(queueName string) ([]string, error) {
	partitions := dq.ring.GetPartitions()
	result := make([]string, 0)
	for _, part := range partitions {
		queue := dq.parts.GetPartition(part).GetDelayQueue()
		res, err := queue.Pop(queueName)
		if err != nil {
			return nil, err
		}
		result = append(result, res...)
	}
	return result, nil
}

func (dq *clusterDelayQueue) PushWithDelay(queueName string, flowId string, delay time.Duration, message []byte) error {
	queue := dq.parts.GetPartition(dq.ring.GetPartition(flowId)).GetDelayQueue()
	return queue.PushWithDelay(queueName, delay, message)
}

type FlowDao interface {
	SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error
	CreateAndSaveFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error)
	GetFlowContext(wfName string, flowId string) (*model.FlowContext, error)
	DeleteFlowContext(wfName string, flowId string) error
}

type clusterFlowDao struct {
	parts *persistence.Partitions
	ring  *Ring
}

var _ FlowDao = new(clusterFlowDao)

func NewFlowDao(parts *persistence.Partitions, ring *Ring) *clusterFlowDao {
	return &clusterFlowDao{
		parts: parts,
		ring:  ring,
	}
}

func (cd *clusterFlowDao) SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error {
	flowDao := cd.parts.GetPartition(cd.ring.GetPartition(flowId)).GetFlowDao()
	return flowDao.SaveFlowContext(wfName, flowId, flowCtx)
}

func (cd *clusterFlowDao) CreateAndSaveFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error) {
	flowDao := cd.parts.GetPartition(cd.ring.GetPartition(flowId)).GetFlowDao()
	return flowDao.CreateAndSaveFlowContext(wFname, flowId, action, dataMap)
}

func (cd *clusterFlowDao) GetFlowContext(wfName string, flowId string) (*model.FlowContext, error) {
	flowDao := cd.parts.GetPartition(cd.ring.GetPartition(flowId)).GetFlowDao()
	return flowDao.GetFlowContext(wfName, flowId)
}
func (cd *clusterFlowDao) DeleteFlowContext(wfName string, flowId string) error {
	flowDao := cd.parts.GetPartition(cd.ring.GetPartition(flowId)).GetFlowDao()
	return flowDao.DeleteFlowContext(wfName, flowId)
}
