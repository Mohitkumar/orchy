package cluster

import (
	"time"

	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/shard"
)

var _ Storage = new(clusterStorage)

type clusterStorage struct {
	shards map[int]*shard.Shard
	ring   *Ring
}

func NewClusterStorage(shards map[int]*shard.Shard, ring *Ring) *clusterStorage {
	return &clusterStorage{
		shards: shards,
		ring:   ring,
	}
}

func (s *clusterStorage) SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error {
	shard := s.shards[s.ring.GetPartition(flowId)].GetStorage()
	return shard.SaveFlowContext(wfName, flowId, flowCtx)
}

func (s *clusterStorage) GetFlowContext(wfName string, flowId string) (*model.FlowContext, error) {
	shard := s.shards[s.ring.GetPartition(flowId)].GetStorage()
	return shard.GetFlowContext(wfName, flowId)
}
func (s *clusterStorage) DeleteFlowContext(wfName string, flowId string) error {
	shard := s.shards[s.ring.GetPartition(flowId)].GetStorage()
	return shard.DeleteFlowContext(wfName, flowId)
}

func (s *clusterStorage) SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, actions []model.ActionExecutionRequest) error {
	shard := s.shards[s.ring.GetPartition(flowId)].GetStorage()
	return shard.SaveFlowContextAndDispatchAction(wfName, flowId, flowCtx, actions)
}

func (s *clusterStorage) Retry(wfName string, flowId string, actionId int, delay time.Duration) error {
	shard := s.shards[s.ring.GetPartition(flowId)].GetStorage()
	return shard.Retry(wfName, flowId, actionId, delay)
}

func (s *clusterStorage) Delay(wfName string, flowId string, actionId int, delay time.Duration) error {
	shard := s.shards[s.ring.GetPartition(flowId)].GetStorage()
	return shard.Delay(wfName, flowId, actionId, delay)
}

func (s *clusterStorage) Timeout(wfName string, flowId string, actionId int, delay time.Duration) error {
	shard := s.shards[s.ring.GetPartition(flowId)].GetStorage()
	return shard.Timeout(wfName, flowId, actionId, delay)
}
