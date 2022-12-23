package cluster

import (
	"sync"
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
)

type ExternalQueue interface {
	Push(action *api.Action) error
	Poll(actionName string, batchSize int) (*api.Actions, error)
}

type Storage interface {
	SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error
	CreateAndSaveFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error)
	GetFlowContext(wfName string, flowId string) (*model.FlowContext, error)
	DeleteFlowContext(wfName string, flowId string) error
	DispatchAction(action *api.Action, actionType string) error
	SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, action *api.Action, actionType string) error
	Retry(action *api.Action, delay time.Duration) error
	Delay(action *api.Action, delay time.Duration) error
	Timeout(action *api.Action, delay time.Duration) error
}

var _ Storage = new(clusterStorage)

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

func (s *clusterStorage) DispatchAction(action *api.Action, actionType string) error {
	shard := s.shards.GetShard(s.ring.GetPartition(action.FlowId))
	return shard.DispatchAction(action, string(actionType))
}
func (s *clusterStorage) SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, action *api.Action, actionType string) error {
	shard := s.shards.GetShard(s.ring.GetPartition(flowId))
	return shard.SaveFlowContextAndDispatchAction(wfName, flowId, flowCtx, action, string(actionType))
}
func (s *clusterStorage) PollAction(actionType string, batchSize int) (*api.Actions, error) {
	var result []*api.Action
	partitions := s.ring.GetPartitions()
	for _, partition := range partitions {
		if len(result) < batchSize {
			numOfItemsToFetch := batchSize - len(result)
			shard := s.shards.GetShard(partition)
			items, err := shard.PollAction(string(actionType), numOfItemsToFetch)
			if err != nil {
				return nil, err
			}
			result = append(result, items.Actions...)
		} else {
			break
		}
	}
	return &api.Actions{Actions: result}, nil
}

func (s *clusterStorage) Retry(action *api.Action, delay time.Duration) error {
	shard := s.shards.GetShard(s.ring.GetPartition(action.FlowId))
	return shard.Retry(action, delay)
}
func (s *clusterStorage) PollRetry(batchSize int) (*api.Actions, error) {
	var result []*api.Action
	partitions := s.ring.GetPartitions()
	for _, partition := range partitions {
		if len(result) < batchSize {
			numOfItemsToFetch := batchSize - len(result)
			shard := s.shards.GetShard(partition)
			items, err := shard.PollRetry(numOfItemsToFetch)
			if err != nil {
				return nil, err
			}
			result = append(result, items.Actions...)
		} else {
			break
		}
	}
	return &api.Actions{Actions: result}, nil
}
func (s *clusterStorage) Delay(action *api.Action, delay time.Duration) error {
	shard := s.shards.GetShard(s.ring.GetPartition(action.FlowId))
	return shard.Delay(action, delay)
}
func (s *clusterStorage) PollDelay(batchSize int) (*api.Actions, error) {
	var result []*api.Action
	partitions := s.ring.GetPartitions()
	for _, partition := range partitions {
		if len(result) < batchSize {
			numOfItemsToFetch := batchSize - len(result)
			shard := s.shards.GetShard(partition)
			items, err := shard.PollDelay(numOfItemsToFetch)
			if err != nil {
				return nil, err
			}
			result = append(result, items.Actions...)
		} else {
			break
		}
	}
	return &api.Actions{Actions: result}, nil
}
func (s *clusterStorage) Timeout(action *api.Action, delay time.Duration) error {
	shard := s.shards.GetShard(s.ring.GetPartition(action.FlowId))
	return shard.Timeout(action, delay)
}
func (s *clusterStorage) PollTimeout(batchSize int) (*api.Actions, error) {
	var result []*api.Action
	partitions := s.ring.GetPartitions()
	for _, partition := range partitions {
		if len(result) < batchSize {
			numOfItemsToFetch := batchSize - len(result)
			shard := s.shards.GetShard(partition)
			items, err := shard.PollTimeout(numOfItemsToFetch)
			if err != nil {
				return nil, err
			}
			result = append(result, items.Actions...)
		} else {
			break
		}
	}
	return &api.Actions{Actions: result}, nil
}
