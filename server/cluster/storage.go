package cluster

import (
	"sync"
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
)

type Storage interface {
	SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error
	CreateAndSaveFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error)
	GetFlowContext(wfName string, flowId string) (*model.FlowContext, error)
	DeleteFlowContext(wfName string, flowId string) error
	DispatchAction(wfName string, flowId string, action *api.Action) error
	SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, action *api.Action) error
	PollAction(actionName string, batchSize int) (*api.Actions, error)
	Retry(req *model.ActionExecutionRequest, delay time.Duration) error
	PollRetry(batch int) (*model.ActionExecutionRequest, error)
	Delay(req *model.ActionExecutionRequest, delay time.Duration) error
	PollDelay(batch int) (*model.ActionExecutionRequest, error)
	Timeout(req *model.ActionExecutionRequest, delay time.Duration) error
	PollTimeout(batch int) (*model.ActionExecutionRequest, error)
}

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

type clusterStorage struct {
	shards *persistence.Shards
	ring   *Ring
	mu     sync.Mutex
}

func NewClusterStorage(shards *persistence.Shards, ring *Ring) *clusterStorage {
	return &clusterStorage{
		shards: shards,
		ring:   ring,
	}
}

func (s *clusterStorage) SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error {
	shard := s.shards.GetShard(s.ring.GetPartition(flowId))
	return shard.SaveFlowContext(wfName, flowId, flowCtx)
}

func (s *clusterStorage) CreateAndSaveFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error) {
	shard := s.shards.GetShard(s.ring.GetPartition(flowId))
	return shard.CreateAndSaveFlowContext(wFname, flowId, action, dataMap)
}

func (s *clusterStorage) GetFlowContext(wfName string, flowId string) (*model.FlowContext, error) {
	shard := s.shards.GetShard(s.ring.GetPartition(flowId))
	return shard.GetFlowContext(wfName, flowId)
}
func (s *clusterStorage) DeleteFlowContext(wfName string, flowId string) error {
	shard := s.shards.GetShard(s.ring.GetPartition(flowId))
	return shard.DeleteFlowContext(wfName, flowId)
}

func (s *clusterStorage) DispatchAction(wfName string, flowId string, action *api.Action) error {

}
func (r *redisShard) SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, action *api.Action) error {

}
func (s *clusterStorage) PollAction(actionName string, batchSize int) (*api.Actions, error) {
	var result []string
	partitions := s.ring.GetPartitions()
	for _, partition := range partitions {
		if len(result) < batchSize {
			numOfItemsToFetch := batchSize - len(result)
			shard := s.shards.GetShard(partition)
			items, err := shard.Pop(queueName, numOfItemsToFetch)
			if err != nil {
				return nil, err
			}
			result = append(result, items...)
		} else {
			break
		}
	}
}

func (r *redisShard) Retry(req *model.ActionExecutionRequest, delay time.Duration) error {

}
func (r *redisShard) PollRetry(batch int) (*model.ActionExecutionRequest, error) {

}
func (r *redisShard) Delay(req *model.ActionExecutionRequest, delay time.Duration) error {

}
func (r *redisShard) PollDelay(batch int) (*model.ActionExecutionRequest, error) {

}
func (r *redisShard) Timeout(req *model.ActionExecutionRequest, delay time.Duration) error {

}
func (r *redisShard) PollTimeout(batch int) (*model.ActionExecutionRequest, error) {

}

func (cq *clusterStorage) Push(queueName string, flowId string, mesage []byte) error {
	queue := cq.parts.GetPartition(cq.ring.GetPartition(flowId)).GetQueue()
	return queue.Push(queueName, mesage)
}

func (cq *clusterStorage) Pop(queueName string, batchSize int) ([]string, error) {
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
