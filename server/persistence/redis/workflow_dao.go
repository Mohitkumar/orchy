package redis

import (
	"context"

	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
)

var _ persistence.WorkflowDao = new(redisWorkflowDao)

const WORKFLOW_DEF string = "WF_DEF"

type redisWorkflowDao struct {
	baseDao
	encoderDecoder util.EncoderDecoder[model.Workflow]
}

func NewRedisWorkflowDao(conf Config) *redisWorkflowDao {
	return &redisWorkflowDao{
		baseDao:        *newBaseDao(conf),
		encoderDecoder: util.NewJsonEncoderDecoder[model.Workflow](),
	}
}

func (rfd *redisWorkflowDao) Save(wf model.Workflow) error {
	key := rfd.baseDao.getNamespaceKey(WORKFLOW_DEF, wf.Name)
	ctx := context.Background()
	data, err := rfd.encoderDecoder.Encode(wf)
	if err != nil {
		return err
	}
	err = rfd.redisClient.Set(ctx, key, data, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rfd *redisWorkflowDao) Delete(name string) error {
	key := rfd.baseDao.getNamespaceKey(WORKFLOW_DEF, name)
	ctx := context.Background()

	err := rfd.redisClient.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rfd *redisWorkflowDao) Get(name string) (*model.Workflow, error) {
	key := rfd.baseDao.getNamespaceKey(WORKFLOW_DEF, name)
	ctx := context.Background()
	val, err := rfd.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	wf, err := rfd.encoderDecoder.Decode([]byte(val))
	if err != nil {
		return nil, err
	}
	return wf, nil
}
