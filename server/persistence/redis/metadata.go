package redis

import (
	"context"

	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

const WORKFLOW_DEF string = "WORKFLOW"
const ACTION_DEF string = "ACTION"

type redisMetadataStorage struct {
	*baseDao
	workflowEncoderDecoder util.EncoderDecoder[model.Workflow]
	actionEencoderDecoder  util.EncoderDecoder[model.ActionDefinition]
}

func NewRedisMetadataStorage(conf Config) *redisMetadataStorage {
	return &redisMetadataStorage{
		baseDao:                newBaseDao(conf),
		workflowEncoderDecoder: util.NewJsonEncoderDecoder[model.Workflow](),
		actionEencoderDecoder:  util.NewJsonEncoderDecoder[model.ActionDefinition](),
	}
}

func (rfd *redisMetadataStorage) SaveWorkflowDefinition(wf model.Workflow) error {
	key := rfd.baseDao.getNamespaceKey(WORKFLOW_DEF, wf.Name)
	ctx := context.Background()
	data, err := rfd.workflowEncoderDecoder.Encode(wf)
	if err != nil {
		return err
	}
	err = rfd.redisClient.Set(ctx, key, data, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rfd *redisMetadataStorage) DeleteWorkflowDefinition(name string) error {
	key := rfd.baseDao.getNamespaceKey(WORKFLOW_DEF, name)
	ctx := context.Background()

	err := rfd.redisClient.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rfd *redisMetadataStorage) GetWorkflowDefinition(name string) (*model.Workflow, error) {
	key := rfd.baseDao.getNamespaceKey(WORKFLOW_DEF, name)
	ctx := context.Background()
	val, err := rfd.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	wf, err := rfd.workflowEncoderDecoder.Decode([]byte(val))
	if err != nil {
		return nil, err
	}
	return wf, nil
}

func (td *redisMetadataStorage) SaveActionDefinition(action model.ActionDefinition) error {
	data, err := td.actionEencoderDecoder.Encode(action)
	if err != nil {
		return err
	}
	key := td.baseDao.getNamespaceKey(ACTION_DEF)
	ctx := context.Background()
	if err := td.baseDao.redisClient.HSet(ctx, key, []string{action.Name, string(data)}).Err(); err != nil {
		logger.Error("error in saving action definition", zap.String("action", action.Name), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (td *redisMetadataStorage) DeleteActionDefinition(action string) error {
	key := td.baseDao.getNamespaceKey(ACTION_DEF)
	ctx := context.Background()
	if err := td.baseDao.redisClient.HDel(ctx, key, action).Err(); err != nil {
		logger.Error("error in deleting action definition", zap.String("action", action), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (td *redisMetadataStorage) GetActionDefinition(action string) (*model.ActionDefinition, error) {
	key := td.baseDao.getNamespaceKey(ACTION_DEF)
	ctx := context.Background()
	actionStr, err := td.baseDao.redisClient.HGet(ctx, key, action).Result()
	if err != nil {
		logger.Error("error in getting action definition", zap.String("action", action), zap.Error(err))
		return nil, persistence.StorageLayerError{}
	}
	return td.actionEencoderDecoder.Decode([]byte(actionStr))
}
