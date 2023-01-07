package persistence

import (
	"fmt"
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/model"
)

type StorageLayerError struct {
	Message string
}

func (e StorageLayerError) Error() string {
	return fmt.Sprintf("storage layer error %s", e.Message)
}

const WF_PREFIX string = "WF_"
const METADATA_CF string = "METADATA_"

type MetadataStorage interface {
	SaveWorkflowDefinition(wf model.Workflow) error
	DeleteWorkflowDefinition(name string) error
	GetWorkflowDefinition(name string) (*model.Workflow, error)
	SaveActionDefinition(action model.ActionDefinition) error
	DeleteActionDefinition(action string) error
	GetActionDefinition(action string) (*model.ActionDefinition, error)
}

type ExternalQueue interface {
	Push(actions []model.ActionExecutionRequest) error
	Poll(actionName string, batchSize int) ([]model.ActionExecutionRequest, error)
}
type Shard interface {
	GetShardId() string
	SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error
	GetFlowContext(wfName string, flowId string) (*model.FlowContext, error)
	DeleteFlowContext(wfName string, flowId string) error

	SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, actions []model.ActionExecutionRequest) error
	PollAction(actionType string, batchSize int) ([]model.ActionExecutionRequest, error)
	Retry(wfName string, flowId string, actionId int, delay time.Duration) error
	PollRetry() ([]model.ActionExecutionRequest, error)
	Delay(wfName string, flowId string, actionId int, delay time.Duration) error
	PollDelay() ([]model.ActionExecutionRequest, error)
	Timeout(wfName string, flowId string, actionId int, delay time.Duration) error
	PollTimeout() ([]model.ActionExecutionRequest, error)
	GetExternalQueue() ExternalQueue
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
