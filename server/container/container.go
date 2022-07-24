package container

import (
	"github.com/mohitkumar/orchy/server/config"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	rd "github.com/mohitkumar/orchy/server/persistence/redis"
	"github.com/mohitkumar/orchy/server/util"
)

type DIContiner struct {
	initialized                  bool
	wfDao                        persistence.WorkflowDao
	flowDao                      persistence.FlowDao
	taskDao                      persistence.TaskDao
	queue                        persistence.Queue
	delayQueue                   persistence.DelayQueue
	taskTimeoutQueue             persistence.DelayQueue
	taskRetryQueue               persistence.DelayQueue
	FlowContextEncDec            util.EncoderDecoder[model.FlowContext]
	ActionExecutionRequestEncDec util.EncoderDecoder[model.ActionExecutionRequest]
	TaskEncDec                   util.EncoderDecoder[model.TaskDef]
}

func (p *DIContiner) setInitialized() {
	p.initialized = true
}

func NewDiContainer() *DIContiner {
	return &DIContiner{
		initialized: false,
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
		d.flowDao = rd.NewRedisFlowDao(*rdConf, d.FlowContextEncDec)
		d.taskDao = rd.NewRedisTaskDao(*rdConf, d.TaskEncDec)

	case config.STORAGE_TYPE_INMEM:

	}
	switch conf.QueueType {
	case config.QUEUE_TYPE_REDIS:
		rdConf := &rd.Config{
			Addrs:     conf.RedisConfig.Addrs,
			Namespace: conf.RedisConfig.Namespace,
		}
		d.queue = rd.NewRedisQueue(*rdConf)
		d.delayQueue = rd.NewRedisDelayQueue(*rdConf)
		d.taskTimeoutQueue = rd.NewRedisDelayQueue(*rdConf)
		d.taskRetryQueue = rd.NewRedisDelayQueue(*rdConf)
	}

}

func (d *DIContiner) GetWorkflowDao() persistence.WorkflowDao {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.wfDao
}

func (d *DIContiner) GetFlowDao() persistence.FlowDao {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.flowDao
}

func (d *DIContiner) GetTaskDao() persistence.TaskDao {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.taskDao
}

func (d *DIContiner) GetQueue() persistence.Queue {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.queue
}

func (d *DIContiner) GetDelayQueue() persistence.DelayQueue {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.delayQueue
}

func (d *DIContiner) GetTaskTimeoutQueue() persistence.DelayQueue {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.taskTimeoutQueue
}

func (d *DIContiner) GetTaskRetryQueue() persistence.DelayQueue {
	if !d.initialized {
		panic("persistence not initalized")
	}
	return d.taskRetryQueue
}
