package persistence

import (
	"fmt"
	"time"

	"github.com/mohitkumar/orchy/server/model"
)

type EmptyQueueError struct {
	QueueName string
}

func (e EmptyQueueError) Error() string {
	return fmt.Sprintf("%s is empty", e.QueueName)
}

type StorageLayerError struct {
	Message string
}

func (e StorageLayerError) Error() string {
	return fmt.Sprintf("storage layer error %s", e.Message)
}

const WF_PREFIX string = "WF_"
const METADATA_CF string = "METADATA_"

type WorkflowDao interface {
	Save(wf model.Workflow) error

	Delete(name string) error

	Get(name string) (*model.Workflow, error)
}

type FlowDao interface {
	SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error
	CreateAndSaveFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error)
	AddActionOutputToFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error)
	GetFlowContext(wfName string, flowId string) (*model.FlowContext, error)
}

type TaskDao interface {
	SaveTask(task model.TaskDef) error
	DeleteTask(task string) error
	GetTask(task string) (*model.TaskDef, error)
}
type Queue interface {
	Push(queueName string, mesage []byte) error
	Pop(queuName string) ([]byte, error)
}

type DelayQueue interface {
	Push(queueName string, mesage []byte) error
	Pop(queueName string) ([]string, error)
	PushWithDelay(queueName string, delay time.Duration, message []byte) error
}
