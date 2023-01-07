package container

import (
	"github.com/mohitkumar/orchy/server/cluster"
	"github.com/mohitkumar/orchy/server/config"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	rd "github.com/mohitkumar/orchy/server/persistence/redis"
	"github.com/mohitkumar/orchy/server/util"
	v8 "rogchap.com/v8go"
)

type DIContiner struct {
	initialized       bool
	metadataStorage   persistence.MetadataStorage
	clusterStorage    cluster.Storage
	stateHandler      *cluster.StateHandlerContainer
	FlowContextEncDec util.EncoderDecoder[model.FlowContext]
	ActionEncDec      util.EncoderDecoder[model.ActionDefinition]
	ring              *cluster.Ring
	shards            *persistence.Shards
	jsvM              *v8.Isolate
}

func (p *DIContiner) setInitialized() {
	p.initialized = true
}

func NewDiContainer(ring *cluster.Ring) *DIContiner {
	return &DIContiner{
		initialized: false,
		ring:        ring,
	}
}

func (d *DIContiner) Init(conf config.Config) {
	defer d.setInitialized()

	switch conf.EncoderDecoderType {
	case config.PROTO_ENCODER_DECODER:
		//proto
	default:
		d.FlowContextEncDec = util.NewJsonEncoderDecoder[model.FlowContext]()
		d.ActionEncDec = util.NewJsonEncoderDecoder[model.ActionDefinition]()
	}

	var externalQueue persistence.ExternalQueue
	switch conf.QueueType {
	case config.QUEUE_TYPE_REDIS:
		rdConf := rd.Config{
			Addrs:     conf.RedisConfig.Addrs,
			Namespace: conf.RedisConfig.Namespace,
		}
		externalQueue = rd.NewRedisQueue(rdConf)
	}
	switch conf.StorageType {
	case config.STORAGE_TYPE_REDIS:
		rdConf := rd.Config{
			Addrs:     conf.RedisConfig.Addrs,
			Namespace: conf.RedisConfig.Namespace,
		}
		d.shards = rd.InitRedisShards(rdConf, externalQueue, d.FlowContextEncDec, conf.RingConfig.PartitionCount)
		d.metadataStorage = rd.NewRedisMetadataStorage(rdConf)
		d.clusterStorage = cluster.NewClusterStorage(d.shards, d.ring)
	case config.STORAGE_TYPE_INMEM:
	}

	d.stateHandler = cluster.NewStateHandlerContainer(d.clusterStorage)
	d.jsvM = v8.NewIsolate()
	d.stateHandler.Init()
}

func (d *DIContiner) GetShards() *persistence.Shards {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.shards
}

func (d *DIContiner) GetMetadataStorage() persistence.MetadataStorage {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.metadataStorage
}

func (d *DIContiner) GetClusterStorage() cluster.Storage {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.clusterStorage
}

func (d *DIContiner) GetStateHandler() *cluster.StateHandlerContainer {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.stateHandler
}

func (d *DIContiner) GetJavaScriptVM() *v8.Isolate {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.jsvM
}
