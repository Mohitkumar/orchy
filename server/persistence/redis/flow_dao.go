package redis

import (
	"context"
	"fmt"

	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

const WORKFLOW_KEY string = "WF"

var _ persistence.FlowDao = new(redisFlowDao)

type redisFlowDao struct {
	baseDao
	encoderDecoder util.EncoderDecoder[model.FlowContext]
}

func NewRedisFlowDao(conf Config, encoderDecoder util.EncoderDecoder[model.FlowContext]) *redisFlowDao {
	return &redisFlowDao{
		baseDao:        *newBaseDao(conf),
		encoderDecoder: encoderDecoder,
	}
}
func (rf *redisFlowDao) CreateAndSaveFlowContext(wFname string, flowId string, action int, input map[string]any) (*model.FlowContext, error) {
	dataMap := make(map[string]any)
	dataMap["input"] = input
	flowCtx := &model.FlowContext{
		Id:            flowId,
		State:         model.RUNNING,
		CurrentAction: action,
		Data:          dataMap,
	}
	if err := rf.SaveFlowContext(wFname, flowId, flowCtx); err != nil {
		return nil, err
	}

	return flowCtx, nil
}

func (rf *redisFlowDao) AddActionOutputToFlowContext(wFname string, flowId string, action int, dataMap map[string]any) (*model.FlowContext, error) {
	flowCtx, err := rf.GetFlowContext(wFname, flowId)
	if err != nil {
		return nil, err
	}
	data := flowCtx.Data
	output := make(map[string]any)
	output["output"] = dataMap
	data[fmt.Sprintf("%d", action)] = util.ConvertMapToStructPb(output)
	if err := rf.SaveFlowContext(wFname, flowId, flowCtx); err != nil {
		return nil, err
	}
	return flowCtx, nil
}

func (rf *redisFlowDao) SaveFlowContext(wfName string, flowId string, flowCtx *model.FlowContext) error {
	key := rf.baseDao.getNamespaceKey(WORKFLOW_KEY, wfName)
	ctx := context.Background()
	data, err := rf.encoderDecoder.Encode(*flowCtx)
	if err != nil {
		return err
	}
	if err := rf.baseDao.redisClient.HSet(ctx, key, []string{flowId, string(data)}).Err(); err != nil {
		logger.Error("error in saving flow context", zap.String("flowName", wfName), zap.String("flowId", flowId), zap.Error(err))
		return persistence.StorageLayerError{}
	}
	return nil
}

func (rf *redisFlowDao) GetFlowContext(wfName string, flowId string) (*model.FlowContext, error) {
	key := rf.baseDao.getNamespaceKey(WORKFLOW_KEY, wfName)
	ctx := context.Background()
	flowCtxStr, err := rf.baseDao.redisClient.HGet(ctx, key, flowId).Result()
	if err != nil {
		logger.Error("error in getting flow context", zap.String("flowName", wfName), zap.String("flowId", flowId), zap.Error(err))
		return nil, persistence.StorageLayerError{}
	}

	flowCtx, err := rf.encoderDecoder.Decode([]byte(flowCtxStr))
	if err != nil {
		return nil, err
	}
	return flowCtx, nil
}
