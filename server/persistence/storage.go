package persistence

import (
	"fmt"

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
