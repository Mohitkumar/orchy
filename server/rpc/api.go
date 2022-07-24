package rpc

import (
	"context"
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"go.uber.org/zap"
)

var _ api.TaskServiceServer = (*grpcServer)(nil)

func (srv *grpcServer) SaveTaskDef(ctx context.Context, req *api.TaskDef) (*api.TaskDefSaveResponse, error) {
	task := &model.TaskDef{
		Name:              req.Name,
		RetryCount:        int(req.RetryCount),
		RetryAfterSeconds: int(req.RetryAfterSeconds),
		RetryPolicy:       model.RetryPolicy(req.RetryPolicy),
		TimeoutSeconds:    int(req.TimeoutSeconds),
	}
	err := srv.TaskDefService.SaveTask(*task)
	if err != nil {
		return &api.TaskDefSaveResponse{
			Status: false,
		}, err
	}
	return &api.TaskDefSaveResponse{Status: true}, nil
}

func (srv *grpcServer) Poll(ctx context.Context, req *api.TaskPollRequest) (*api.Task, error) {
	task, err := srv.TaskService.Poll(req.TaskType)
	if err != nil {
		switch err.(type) {
		case persistence.StorageLayerError:
			return nil, &api.StorageLayerError{}
		case persistence.EmptyQueueError:
			return nil, &api.PollError{QueueName: req.TaskType}
		}
	}
	return task, nil
}

func (srv *grpcServer) Push(ctx context.Context, req *api.TaskResult) (*api.TaskResultPushResponse, error) {
	err := srv.TaskService.Push(req)
	if err != nil {
		return &api.TaskResultPushResponse{
			Status: false,
		}, err
	}
	return &api.TaskResultPushResponse{Status: true}, nil
}

func (srv *grpcServer) PollStream(req *api.TaskPollRequest, stream api.TaskService_PollStreamServer) error {
	timer := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-timer.C:
			res, err := srv.TaskService.Poll(req.TaskType)
			if err != nil {
				switch err.(type) {
				case persistence.StorageLayerError:
					logger.Error("storage layer error while polling ", zap.Error(err))
				case persistence.EmptyQueueError:
					logger.Debug("task queue is empty", zap.String("queue", req.TaskType))
				}
			} else {
				if err = stream.Send(res); err != nil {
					logger.Error("error sending stream response", zap.Error(err))
				}
			}
		}
	}
}
