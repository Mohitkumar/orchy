package cluster

import (
	"strconv"
	"sync"
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/config"
	"github.com/mohitkumar/orchy/server/metadata"
	"github.com/mohitkumar/orchy/server/model"
	rd "github.com/mohitkumar/orchy/server/persistence/redis"
	"github.com/mohitkumar/orchy/server/shard"
	"github.com/mohitkumar/orchy/server/shard/executor"
	"github.com/mohitkumar/orchy/server/util"
	v8 "rogchap.com/v8go"
)

type Cluster struct {
	storage         Storage
	ring            *Ring
	membership      *Membership
	shards          map[int]*shard.Shard
	stateHandler    *StateHandlerContainer
	metadataService metadata.MetadataService
	jsvm            *v8.Isolate
	mu              sync.Mutex
}

func NewCluster(conf config.Config, flowExecutionChannel chan<- model.FlowExecutionRequest, wg *sync.WaitGroup) *Cluster {
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
	var metadataStorage metadata.MetadataStorage
	switch conf.StorageType {
	case config.STORAGE_TYPE_REDIS:
		rdConf := rd.Config{
			Addrs:     conf.RedisConfig.Addrs,
			Namespace: conf.RedisConfig.Namespace,
		}
		metadataStorage = rd.NewRedisMetadataStorage(rdConf)
	case config.STORAGE_TYPE_INMEM:
	}
	jsvm := v8.NewIsolate()
	metadataService := metadata.NewMetadataService(metadataStorage, jsvm)
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

		sh.RegisterExecutor("user-action", executor.NewUserActionExecutor(shardId, shardStorage, metadataService, externalQueue, wg))
		sh.RegisterExecutor("system-action", executor.NewSystemActionExecutor(shardId, shardStorage, flowExecutionChannel, wg))
		sh.RegisterExecutor("delay", executor.NewDelayExecutor(shardId, shardStorage, flowExecutionChannel, wg))
		sh.RegisterExecutor("retry", executor.NewRetryExecutor(shardId, shardStorage, flowExecutionChannel, wg))
		sh.RegisterExecutor("timeout", executor.NewTimeoutExecutor(shardId, shardStorage, flowExecutionChannel, wg))
		shards[i] = sh
	}

	clusterStorage := NewClusterStorage(shards, ring)
	stateHandler := NewStateHandlerContainer(clusterStorage)

	return &Cluster{
		storage:         clusterStorage,
		metadataService: metadataService,
		ring:            ring,
		membership:      membership,
		stateHandler:    stateHandler,
		shards:          make(map[int]*shard.Shard),
		jsvm:            jsvm,
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

func (c *Cluster) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, sh := range c.shards {
		sh.Stop()
	}
	return nil
}

func (c *Cluster) GetStorage() Storage {
	return c.storage
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

func (c *Cluster) GetMetadataService() metadata.MetadataService {
	return c.metadataService
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

func (c *Cluster) GetJsVM() *v8.Isolate {
	return c.jsvm
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
