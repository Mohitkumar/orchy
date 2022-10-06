package rpc

import (
	"context"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
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

func (srv *grpcServer) Poll(ctx context.Context, req *api.TaskPollRequest) (*api.Tasks, error) {
	task, err := srv.TaskService.Poll(req.TaskType, int(req.BatchSize))
	if err != nil {
		switch err.(type) {
		case persistence.StorageLayerError:
			return nil, &api.StorageLayerError{}
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

func (s *grpcServer) GetServers(ctx context.Context, req *api.GetServersRequest) (*api.GetServersResponse, error) {
	servers, err := s.GetServerer.GetServers()
	if err != nil {
		return nil, err
	}
	return &api.GetServersResponse{Servers: servers}, nil
}
