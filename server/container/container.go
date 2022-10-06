package container

import (
	"github.com/mohitkumar/orchy/server/cluster"
	"github.com/mohitkumar/orchy/server/config"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	rd "github.com/mohitkumar/orchy/server/persistence/redis"
	"github.com/mohitkumar/orchy/server/util"
)

type DIContiner struct {
	initialized                  bool
	wfDao                        persistence.WorkflowDao
	taskDao                      persistence.TaskDao
	flowDao                      cluster.FlowDao
	stateHandler                 *cluster.StateHandlerContainer
	queue                        cluster.Queue
	delayQueue                   cluster.DelayQueue
	taskTimeoutQueue             cluster.DelayQueue
	taskRetryQueue               cluster.DelayQueue
	FlowContextEncDec            util.EncoderDecoder[model.FlowContext]
	ActionExecutionRequestEncDec util.EncoderDecoder[model.ActionExecutionRequest]
	TaskEncDec                   util.EncoderDecoder[model.TaskDef]
	ring                         *cluster.Ring
	memebership                  *cluster.Membership
}

func (p *DIContiner) setInitialized() {
	p.initialized = true
}

func NewDiContainer(memberShip *cluster.Membership, ring *cluster.Ring) *DIContiner {
	return &DIContiner{
		initialized: false,
		memebership: memberShip,
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
		d.ActionExecutionRequestEncDec = util.NewJsonEncoderDecoder[model.ActionExecutionRequest]()
		d.TaskEncDec = util.NewJsonEncoderDecoder[model.TaskDef]()
	}

	switch conf.StorageType {
	case config.STORAGE_TYPE_REDIS:
		rdConf := &rd.Config{
			Addrs:     conf.RedisConfig.Addrs,
			Namespace: conf.RedisConfig.Namespace,
		}
		d.wfDao = rd.NewRedisWorkflowDao(*rdConf)
		d.flowDao = cluster.NewFlowDao(rd.NewRedisFlowDao(*rdConf, d.FlowContextEncDec), d.ring)
		d.taskDao = rd.NewRedisTaskDao(*rdConf, d.TaskEncDec)

	case config.STORAGE_TYPE_INMEM:

	}
	switch conf.QueueType {
	case config.QUEUE_TYPE_REDIS:
		rdConf := &rd.Config{
			Addrs:     conf.RedisConfig.Addrs,
			Namespace: conf.RedisConfig.Namespace,
		}
		d.queue = cluster.NewQueue(rd.NewRedisQueue(*rdConf), d.memebership, d.ring)
		d.delayQueue = cluster.NewDelayQueue(rd.NewRedisDelayQueue(*rdConf), d.memebership, d.ring)
		d.taskTimeoutQueue = cluster.NewDelayQueue(rd.NewRedisDelayQueue(*rdConf), d.memebership, d.ring)
		d.taskRetryQueue = cluster.NewDelayQueue(rd.NewRedisDelayQueue(*rdConf), d.memebership, d.ring)
	}
	d.stateHandler = cluster.NewStateHandlerContainer(d.flowDao)
	d.stateHandler.Init()
}

func (d *DIContiner) GetWorkflowDao() persistence.WorkflowDao {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.wfDao
}

func (d *DIContiner) GetFlowDao() cluster.FlowDao {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.flowDao
}

func (d *DIContiner) GetStateHandler() *cluster.StateHandlerContainer {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.stateHandler
}

func (d *DIContiner) GetTaskDao() persistence.TaskDao {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.taskDao
}

func (d *DIContiner) GetQueue() cluster.Queue {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.queue
}

func (d *DIContiner) GetDelayQueue() cluster.DelayQueue {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.delayQueue
}

func (d *DIContiner) GetTaskTimeoutQueue() cluster.DelayQueue {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.taskTimeoutQueue
}

func (d *DIContiner) GetTaskRetryQueue() cluster.DelayQueue {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.taskRetryQueue
}
