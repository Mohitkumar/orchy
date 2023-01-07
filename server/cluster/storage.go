package cluster

import (
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
	GetFlowContext(wfName string, flowId string) (*model.FlowContext, error)
	DeleteFlowContext(wfName string, flowId string) error

	SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, actions []model.ActionExecutionRequest) error
	Retry(wfName string, flowId string, actionId int, delay time.Duration) error
	Delay(wfName string, flowId string, actionId int, delay time.Duration) error
	Timeout(wfName string, flowId string, actionId int, delay time.Duration) error
}

var _ Storage = new(clusterStorage)

type clusterStorage struct {
	shards *persistence.Shards
	ring   *Ring
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

func (s *clusterStorage) GetFlowContext(wfName string, flowId string) (*model.FlowContext, error) {
	shard := s.shards.GetShard(s.ring.GetPartition(flowId))
	return shard.GetFlowContext(wfName, flowId)
}
func (s *clusterStorage) DeleteFlowContext(wfName string, flowId string) error {
	shard := s.shards.GetShard(s.ring.GetPartition(flowId))
	return shard.DeleteFlowContext(wfName, flowId)
}

func (s *clusterStorage) SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, actions []model.ActionExecutionRequest) error {
	shard := s.shards.GetShard(s.ring.GetPartition(flowId))
	return shard.SaveFlowContextAndDispatchAction(wfName, flowId, flowCtx, actions)
}

func (s *clusterStorage) Retry(wfName string, flowId string, actionId int, delay time.Duration) error {
	shard := s.shards.GetShard(s.ring.GetPartition(flowId))
	return shard.Retry(wfName, flowId, actionId, delay)
}

func (s *clusterStorage) Delay(wfName string, flowId string, actionId int, delay time.Duration) error {
	shard := s.shards.GetShard(s.ring.GetPartition(flowId))
	return shard.Delay(wfName, flowId, actionId, delay)
}

func (s *clusterStorage) Timeout(wfName string, flowId string, actionId int, delay time.Duration) error {
	shard := s.shards.GetShard(s.ring.GetPartition(flowId))
	return shard.Timeout(wfName, flowId, actionId, delay)
}
