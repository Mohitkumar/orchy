package redis

import (
	"context"

	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

const TASK_KEY = "TASK_DEF"

var _ persistence.TaskDao = new(redisTaskDao)

type redisTaskDao struct {
	baseDao
	encoderDecoder util.EncoderDecoder[model.TaskDef]
}

func NewRedisTaskDao(conf Config, encoderDecoder util.EncoderDecoder[model.TaskDef]) *redisTaskDao {
	return &redisTaskDao{
		baseDao:        *newBaseDao(conf),
		encoderDecoder: encoderDecoder,
	}
}

func (td *redisTaskDao) SaveTask(task model.TaskDef) error {
	data, err := td.encoderDecoder.Encode(task)
	if err != nil {
		return err
	}
	key := td.baseDao.getNamespaceKey(TASK_KEY)
	ctx := context.Background()
	if err := td.baseDao.redisClient.HSet(ctx, key, []string{task.Name, string(data)}).Err(); err != nil {
		logger.Error("error in saving task definition", zap.String("taskName", task.Name), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (td *redisTaskDao) DeleteTask(task string) error {
	key := td.baseDao.getNamespaceKey(TASK_KEY)
	ctx := context.Background()
	if err := td.baseDao.redisClient.HDel(ctx, key, task).Err(); err != nil {
		logger.Error("error in deleting task definition", zap.String("taskName", task), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (td *redisTaskDao) GetTask(task string) (*model.TaskDef, error) {
	key := td.baseDao.getNamespaceKey(TASK_KEY)
	ctx := context.Background()
	taskStr, err := td.baseDao.redisClient.HGet(ctx, key, task).Result()
	if err != nil {
		logger.Error("error in getting task definition", zap.String("taskName", task), zap.Error(err))
		return nil, persistence.StorageLayerError{}
	}
	return td.encoderDecoder.Decode([]byte(taskStr))
}
