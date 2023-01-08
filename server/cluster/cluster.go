package cluster

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
)

type Cluster struct {
	storage     Storage
	ring        *Ring
	membership  *Membership
	flowService *flow.FlowService
	shards      map[int]persistence.Shard
	mu          sync.Mutex
}

func NewCluster(config Config, storage Storage, flowService *flow.FlowService) *Cluster {
	ring := NewRing(config.PartitionCount)
	membership, err := New(ring, config)
	if err != nil {
		panic("can not start cluster")
	}
	return &Cluster{
		ring:        ring,
		membership:  membership,
		storage:     storage,
		flowService: flowService,
		shards:      make(map[int]persistence.Shard),
	}
}

func (c *Cluster) GetShard(shardId int) persistence.Shard {
	c.mu.Lock()
	defer c.mu.Unlock()
	shard := c.shards[shardId]
	return shard
}

func (c *Cluster) Start() {

}

func (c *Cluster) GetStorage() Storage {
	return c.storage
}

func (c *Cluster) GetClusterRefersher() *Membership {
	return c.membership
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
