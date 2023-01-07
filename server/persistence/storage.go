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

type Shard interface {
	GetShardId() string
	SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error
	GetFlowContext(wfName string, flowId string) (*model.FlowContext, error)
	DeleteFlowContext(wfName string, flowId string) error

	SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, actions []model.ActionExecutionRequest) error
	PollAction(actionType string, batchSize int) ([]string, error)
	Retry(wfName string, flowId string, actionId int, delay time.Duration) error
	PollRetry() ([]string, error)
	Delay(wfName string, flowId string, actionId int, delay time.Duration) error
	PollDelay() ([]string, error)
	Timeout(wfName string, flowId string, actionId int, delay time.Duration) error
	PollTimeout() ([]string, error)
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
