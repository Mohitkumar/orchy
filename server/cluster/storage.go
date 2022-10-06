package cluster

import (
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
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
	redisQueue       persistence.Queue
	membership       *Membership
	ring             *Ring
	mu               sync.Mutex
	currentPartition uint64
}

func newQueue(queue persistence.Queue, membership *Membership, ring *Ring) *clusterQueue {
	return &clusterQueue{
		redisQueue:       queue,
		membership:       membership,
		ring:             ring,
		currentPartition: 0,
	}
}

func (cq *clusterQueue) Push(queueName string, flowId string, mesage []byte) error {
	partition := strconv.Itoa(cq.ring.GetPartition(flowId))
	return cq.redisQueue.Push(queueName, partition, mesage)
}

func (cq *clusterQueue) Pop(queueName string, batchSize int) ([]string, error) {
	result := make([]string, 0)
	for len(result) < batchSize {
		partition := cq.getNextPartition()
		numOfItemsToFetch := batchSize - len(result)
		items, err := cq.redisQueue.Pop(queueName, strconv.Itoa(partition), numOfItemsToFetch)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}
	return result, nil
}

func (cq *clusterQueue) getNextPartition() int {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	partitions := cq.ring.GetPartitions(cq.membership.GetLocalMemebr())
	partitions = sort.IntSlice(partitions)
	idx := sort.Search(len(partitions), func(i int) bool {
		return partitions[i] > int(cq.currentPartition)
	})
	if idx >= len(partitions) {
		idx = 0
	}
	nextPartition := partitions[idx]
	atomic.StoreUint64(&cq.currentPartition, uint64(nextPartition))
	return nextPartition
}

type DelayQueue interface {
	Push(queueName string, flowId string, mesage []byte) error
	Pop(queueName string) ([]string, error)
	PushWithDelay(queueName string, flowId string, delay time.Duration, message []byte) error
}

type clusterDelayQueue struct {
	redisQueue persistence.DelayQueue
	membership *Membership
	ring       *Ring
}

var _ DelayQueue = new(clusterDelayQueue)

func newDelayQueue(queue persistence.DelayQueue, membership *Membership, ring *Ring) *clusterDelayQueue {
	return &clusterDelayQueue{
		redisQueue: queue,
		membership: membership,
		ring:       ring,
	}
}
func (dq *clusterDelayQueue) Push(queueName string, flowId string, mesage []byte) error {
	partition := strconv.Itoa(dq.ring.GetPartition(flowId))
	return dq.redisQueue.Push(queueName, partition, mesage)
}

func (dq *clusterDelayQueue) Pop(queueName string) ([]string, error) {
	partitions := dq.ring.GetPartitions(dq.membership.GetLocalMemebr())
	result := make([]string, 0)
	for part := range partitions {
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

func newFlowDao(dao persistence.FlowDao, ring *Ring) *clusterFlowDao {
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

type ClusterStore struct {
	Queue      Queue
	DelayQueue DelayQueue
	FlowDao    FlowDao
	membership *Membership
	ring       *Ring
}

func NewClusterStore(ring *Ring, memebership *Membership,
	queue persistence.Queue, delayQueue persistence.DelayQueue, flowDao persistence.FlowDao) *ClusterStore {
	return &ClusterStore{
		Queue:      newQueue(queue, memebership, ring),
		DelayQueue: newDelayQueue(delayQueue, memebership, ring),
		FlowDao:    newFlowDao(flowDao, ring),
		membership: memebership,
		ring:       ring,
	}
}
