package cassandra

import (
	"fmt"

	"github.com/mohitkumar/orchy/model"
	"github.com/mohitkumar/orchy/persistence"
	"github.com/mohitkumar/orchy/util"
)

type cassanraMetadataStorage struct {
	*baseDao
	workflowEncoderDecoder util.EncoderDecoder[model.Workflow]
	actionEencoderDecoder  util.EncoderDecoder[model.ActionDefinition]
}

func NewCassandraMetadataStorage(conf Config) *cassanraMetadataStorage {
	return &cassanraMetadataStorage{
		baseDao:                NewBaseDao(conf),
		workflowEncoderDecoder: util.NewJsonEncoderDecoder[model.Workflow](),
		actionEencoderDecoder:  util.NewJsonEncoderDecoder[model.ActionDefinition](),
	}
}

func (rfd *cassanraMetadataStorage) SaveWorkflowDefinition(wf model.Workflow) error {
	data, err := rfd.workflowEncoderDecoder.Encode(wf)
	if err != nil {
		return err
	}
	if err := rfd.Session.Query("CREATE TABLE IF NOT EXIST WORKFLOW(name text, definition text)").Exec(); err != nil {
		return persistence.StorageLayerError{err.Error()}
	}
	if err := rfd.Session.Query("INSERT INTO WORKFLOW(name, definition) VALUES(?,?)", wf.Name, data).Exec(); err != nil {
		return persistence.StorageLayerError{err.Error()}
	}
	return nil
}

func (rfd *cassanraMetadataStorage) DeleteWorkflowDefinition(name string) error {
	if err := rfd.Session.Query("DELETE FROM WORKFLOW WHERE name=?", name).Exec(); err != nil {
		return persistence.StorageLayerError{err.Error()}
	}
	return nil
}

func (rfd *cassanraMetadataStorage) GetWorkflowDefinition(name string) (*model.Workflow, error) {
	var workflows []*model.Workflow
	itr := rfd.Session.Query("select *  FROM WORKFLOW WHERE name=?", name).Iter()
	m := map[string]interface{}{}
	for itr.MapScan(m) {
		wf, err := rfd.workflowEncoderDecoder.Decode([]byte(m["definition"].(string)))
		if err != nil {
			continue
		}
		workflows = append(workflows, wf)
	}
	if len(workflows) == 0 {
		return nil, fmt.Errorf("not found")
	}
	return workflows[0], nil
}

func (td *cassanraMetadataStorage) SaveActionDefinition(action model.ActionDefinition) error {
	data, err := td.actionEencoderDecoder.Encode(action)
	if err != nil {
		return err
	}
	if err := td.Session.Query("CREATE TABLE IF NOT EXIST ACTION(name text, definition text)").Exec(); err != nil {
		return persistence.StorageLayerError{err.Error()}
	}
	if err := td.Session.Query("INSERT INTO ACTION(name, definition) VALUES(?,?)", action.Name, data).Exec(); err != nil {
		return persistence.StorageLayerError{err.Error()}
	}

	return nil
}

func (td *cassanraMetadataStorage) DeleteActionDefinition(action string) error {
	if err := td.Session.Query("DELETE FROM ACTION WHERE name=?", action).Exec(); err != nil {
		return persistence.StorageLayerError{err.Error()}
	}
	return nil
}

func (td *cassanraMetadataStorage) GetActionDefinition(action string) (*model.ActionDefinition, error) {
	var actions []*model.ActionDefinition
	itr := td.Session.Query("select *  FROM ACTION WHERE name=?", action).Iter()
	m := map[string]interface{}{}
	for itr.MapScan(m) {
		act, err := td.actionEencoderDecoder.Decode([]byte(m["definition"].(string)))
		if err != nil {
			continue
		}
		actions = append(actions, act)
	}
	if len(actions) == 0 {
		return nil, fmt.Errorf("not found")
	}
	return actions[0], nil
}
