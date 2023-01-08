package cluster

import (
	"strconv"
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/config"
	"github.com/mohitkumar/orchy/server/model"
	rd "github.com/mohitkumar/orchy/server/persistence/redis"
	"github.com/mohitkumar/orchy/server/shard"
	"github.com/mohitkumar/orchy/server/shard/executor"
	"github.com/mohitkumar/orchy/server/util"
)

type Cluster struct {
	storage      Storage
	ring         *Ring
	membership   *Membership
	shards       map[int]*shard.Shard
	stateHandler *StateHandlerContainer
	mu           sync.Mutex
}

func NewCluster(conf config.Config, wg *sync.WaitGroup) *Cluster {
	cluserConfig := conf.ClusterConfig
	ring := NewRing(cluserConfig.PartitionCount)
	membership, err := New(ring, cluserConfig)
	if err != nil {
		panic("can not start cluster")
	}
	shards := make(map[int]*shard.Shard)
	var flowCtxEncoder util.EncoderDecoder[model.FlowContext]
	switch conf.EncoderDecoderType {
	case config.PROTO_ENCODER_DECODER:
		//proto
	default:
		flowCtxEncoder = util.NewJsonEncoderDecoder[model.FlowContext]()
	}
	for i := 0; i < cluserConfig.PartitionCount; i++ {
		shardId := strconv.FormatInt(int64(i), 10)
		var shardStorage shard.Storage
		var externalQueue shard.ExternalQueue
		switch conf.StorageType {
		case config.STORAGE_TYPE_REDIS:
			rdConf := rd.Config{
				Addrs:     conf.RedisConfig.Addrs,
				Namespace: conf.RedisConfig.Namespace,
			}
			shardStorage = rd.NewRedisStorage(rdConf, flowCtxEncoder, shardId)

		case config.STORAGE_TYPE_INMEM:
		}
		switch conf.QueueType {
		case config.QUEUE_TYPE_REDIS:
			rdConf := rd.Config{
				Addrs:     conf.RedisConfig.Addrs,
				Namespace: conf.RedisConfig.Namespace,
			}
			externalQueue = rd.NewRedisQueue(rdConf, shardId)
		}
		sh := shard.NewShard(shardId, externalQueue, shardStorage)

		sh.RegisterExecutor("user-action", executor.NewUserActionExecutor(shardId, shardStorage, externalQueue, wg))
		sh.RegisterExecutor("system-action", executor.NewSystemActionExecutor(shardId, shardStorage, externalQueue, wg))
		sh.RegisterExecutor("delay", executor.NewDelayExecutor(shardId, shardStorage, externalQueue, wg))
		sh.RegisterExecutor("retry", executor.NewRetryExecutor(shardId, shardStorage, externalQueue, wg))
		sh.RegisterExecutor("timeout", executor.NewTimeoutExecutor(shardId, shardStorage, externalQueue, wg))
		shards[i] = sh
	}

	clusterStorage := NewClusterStorage(shards, ring)
	stateHandler := NewStateHandlerContainer(clusterStorage)
	return &Cluster{
		storage:      clusterStorage,
		ring:         ring,
		membership:   membership,
		stateHandler: stateHandler,
		shards:       make(map[int]*shard.Shard),
	}
}

func (c *Cluster) GetShard(shardId int) *shard.Shard {
	c.mu.Lock()
	defer c.mu.Unlock()
	shard := c.shards[shardId]
	return shard
}

func (c *Cluster) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, sh := range c.shards {
		sh.Start()
	}
}

func (c *Cluster) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, sh := range c.shards {
		sh.Stop()
	}
}

func (c *Cluster) GetStorage() Storage {
	return c.storage
}

func (c *Cluster) GetClusterRefersher() *Membership {
	return c.membership
}

func (c *Cluster) GetServerer() *Ring {
	return c.ring
}

func (c *Cluster) GetStateHandler() *StateHandlerContainer {
	return c.stateHandler
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
