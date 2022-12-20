package persistence

import (
	"sync"
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/model"
)

type MetadataStorage interface {
	SaveWorkflowDefinition(wf model.Workflow) error
	DeleteWorkflowDefinition(name string) error
	GetWorkflowDefinition(name string) (*model.Workflow, error)
	SaveActionDefinition(action model.ActionDefinition) error
	DeleteActionDefinition(action string) error
	GetActionDefinition(action string) (*model.ActionDefinition, error)
}

type Shard interface {
	SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error
	CreateAndSaveFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error)
	GetFlowContext(wfName string, flowId string) (*model.FlowContext, error)
	DeleteFlowContext(wfName string, flowId string) error
	DispatchAction(action *api.Action, actionType string) error
	SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, action *api.Action, actionType string) error
	PollAction(actionType string, batchSize int) (*api.Actions, error)
	Retry(action *api.Action, delay time.Duration) error
	PollRetry(batch int) (*api.Actions, error)
	Delay(action *api.Action, delay time.Duration) error
	PollDelay(batch int) (*api.Actions, error)
	Timeout(action *api.Action, delay time.Duration) error
	PollTimeout(batch int) (*api.Actions, error)
}

type ExternalStorage interface {
	Push(queueName string, mesage []byte) error
	Pop(queuName string, batchSize int) ([]string, error)
}

type Shards struct {
	Shards map[int]Shard
	mu     sync.Mutex
}

func (p *Shards) GetShard(shardId int) Shard {
	p.mu.Lock()
	defer p.mu.Unlock()
	shard := p.Shards[shardId]
	return shard
}