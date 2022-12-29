package persistence

import (
	"fmt"
	"sync"
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
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
	CreateAndSaveFlowContext(wFname string, flowId string, actionIds map[int]bool, dataMap map[string]any) (*model.FlowContext, error)
	GetFlowContext(wfName string, flowId string) (*model.FlowContext, error)
	DeleteFlowContext(wfName string, flowId string) error
	DispatchAction(action []*api.Action) error
	SaveFlowContextAndDispatchAction(wfName string, flowId string, flowCtx *model.FlowContext, action []*api.Action) error
	PollAction(actionType string, batchSize int) (*api.Actions, error)
	Retry(action *api.Action, delay time.Duration) error
	PollRetry() (*api.Actions, error)
	Delay(action *api.Action, delay time.Duration) error
	PollDelay() (*api.Actions, error)
	Timeout(action *api.Action, delay time.Duration) error
	PollTimeout() (*api.Actions, error)
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
