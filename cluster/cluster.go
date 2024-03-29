package cluster

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/config"
	"github.com/mohitkumar/orchy/metadata"
	"github.com/mohitkumar/orchy/model"
	rd "github.com/mohitkumar/orchy/persistence/redis"
	"github.com/mohitkumar/orchy/shard"
	"github.com/mohitkumar/orchy/shard/executor"
	"github.com/mohitkumar/orchy/util"
)

type Cluster struct {
	ring       *Ring
	membership *Membership
	shards     map[int]*shard.Shard
	mu         sync.Mutex
}

func NewCluster(conf config.Config, metadataService metadata.MetadataService, wg *sync.WaitGroup) *Cluster {
	cluserConfig := conf.ClusterConfig
	batchSize := conf.BatchSize
	shards := make(map[int]*shard.Shard)
	var flowCtxEncoder util.EncoderDecoder[model.FlowContext]
	switch conf.EncoderDecoderType {
	case config.PROTO_ENCODER_DECODER:
		//proto
	default:
		flowCtxEncoder = util.NewJsonEncoderDecoder[model.FlowContext]()
	}
	rdConf := rd.Config{
		Addrs:     conf.RedisConfig.Addrs,
		Namespace: conf.RedisConfig.Namespace,
		Password:  conf.RedisConfig.Password,
		PoolSize:  100,
	}
	baseDao := rd.NewBaseDao(rdConf)
	for i := 0; i < cluserConfig.PartitionCount; i++ {
		shardId := strconv.FormatInt(int64(i), 10)
		var shardStorage shard.Storage
		var externalQueue shard.ExternalQueue
		switch conf.StorageType {
		case config.STORAGE_TYPE_REDIS:
			shardStorage = rd.NewRedisStorage(baseDao, flowCtxEncoder, shardId)
		case config.STORAGE_TYPE_INMEM:
		}
		switch conf.QueueType {
		case config.QUEUE_TYPE_REDIS:
			externalQueue = rd.NewRedisQueue(baseDao, shardId)
		}
		stateHandler := shard.NewStateHandlerContainer(shardStorage)
		container := shard.NewFlowStateMachineContainer(shardStorage, metadataService)
		engine := shard.NewFlowEngine(shardStorage, metadataService, container)
		sh := shard.NewShard(shardId, externalQueue, shardStorage, engine, stateHandler)

		sh.RegisterExecutor("user-action", executor.NewUserActionExecutor(shardId, shardStorage, metadataService, externalQueue, batchSize, wg))
		sh.RegisterExecutor("system-action", executor.NewSystemActionExecutor(shardId, shardStorage, engine, batchSize, wg))
		sh.RegisterExecutor("delay", executor.NewDelayExecutor(shardId, shardStorage, engine, wg))
		sh.RegisterExecutor("retry", executor.NewRetryExecutor(shardId, shardStorage, engine, wg))
		sh.RegisterExecutor("timeout", executor.NewTimeoutExecutor(shardId, shardStorage, engine, wg))
		shards[i] = sh
	}

	ring := NewRing(cluserConfig.PartitionCount)

	c := &Cluster{
		ring:   ring,
		shards: shards,
	}
	ring.SetRebalancer(c.Rebalance)
	membership, err := NewMemberShip(ring, cluserConfig)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	c.membership = membership
	return c
}

func (c *Cluster) GetShard(shardId int) *shard.Shard {
	c.mu.Lock()
	defer c.mu.Unlock()
	shard := c.shards[shardId]
	return shard
}

func (c *Cluster) Rebalance(partitions []int) {
	for id, shard := range c.shards {
		if contains(partitions, id) {
			shard.Start()
		} else {
			shard.Stop()
		}
	}
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (c *Cluster) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, sh := range c.shards {
		sh.Stop()
	}
	return nil
}

func (c *Cluster) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, sh := range c.shards {
		sh.Start()
	}
	return nil
}

func (c *Cluster) Poll(queuName string, batchSize int) (*api.Actions, error) {
	result := make([]*api.Action, 0)
	for _, part := range c.ring.GetPartitions() {
		if len(result) < batchSize {
			numOfItemsToFetch := batchSize - len(result)
			actions, err := c.shards[part].GetExternalQueue().Poll(queuName, numOfItemsToFetch)
			if err != nil {
				continue
			}
			result = append(result, actions.Actions...)
		} else {
			break
		}
	}
	return &api.Actions{Actions: result}, nil
}

func (c *Cluster) GetServerer() *Ring {
	return c.ring
}

func (c *Cluster) ExecuteAction(wfName string, flowId string, event string, actionId int, data map[string]any) {
	shard := c.shards[c.ring.GetPartition(flowId)]
	shard.GetEngine().ExecuteAction(wfName, flowId, event, actionId, data)
}

func (c *Cluster) RetryAction(wfName string, flowId string, actionName string, actionId int, reason string) {
	shard := c.shards[c.ring.GetPartition(flowId)]
	shard.GetEngine().RetryFailedAction(wfName, flowId, actionName, actionId)
}

func (c *Cluster) Init(wfName string, input map[string]any) (string, error) {
	flowId := uuid.New().String()
	shard := c.shards[c.ring.GetPartition(flowId)]
	err := shard.GetEngine().Init(wfName, flowId, input)
	if err != nil {
		return "", nil
	}
	return flowId, nil
}

func (c *Cluster) ExecuteResume(wfName string, flowId string, event string) {
	shard := c.shards[c.ring.GetPartition(flowId)]
	shard.GetEngine().ExecuteResume(wfName, flowId, event)
}

func (c *Cluster) GetFlow(wfName string, flowId string) (*model.FlowContext, error) {
	shard := c.shards[c.ring.GetPartition(flowId)]
	return shard.GetEngine().GetFlow(wfName, flowId)
}

func (c *Cluster) MarkPaused(wfName string, flowId string) {
	shard := c.shards[c.ring.GetPartition(flowId)]
	shard.GetEngine().MarkPaused(wfName, flowId)
}

func (c *Cluster) ExecuteResumeAfterWait(wfName string, flowId string, event string) error {
	shard := c.shards[c.ring.GetPartition(flowId)]
	return shard.GetEngine().ExecuteResumeAfterWait(wfName, flowId, event)
}
func (c *Cluster) Timeout(wfName string, flowId string, actionName string, actionId int, delay time.Duration) error {
	shard := c.shards[c.ring.GetPartition(flowId)]
	return shard.GetStorage().Timeout(wfName, flowId, actionName, actionId, delay)
}
