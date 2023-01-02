package rpc

import (
	"context"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"github.com/mohitkumar/orchy/server/persistence"
	"go.uber.org/zap"
)

var _ api.ActionServiceServer = (*grpcServer)(nil)

func (srv *grpcServer) SaveActionDefinition(ctx context.Context, req *api.ActionDefinition) (*api.ActionDefinitionSaveResponse, error) {
	action := &model.ActionDefinition{
		Name:              req.Name,
		RetryCount:        int(req.RetryCount),
		RetryAfterSeconds: int(req.RetryAfterSeconds),
		RetryPolicy:       model.RetryPolicy(req.RetryPolicy),
		TimeoutSeconds:    int(req.TimeoutSeconds),
	}
	err := srv.ActionDefinitionService.SaveActionDefinition(*action)
	if err != nil {
		return &api.ActionDefinitionSaveResponse{
			Status: false,
		}, err
	}
	return &api.ActionDefinitionSaveResponse{Status: true}, nil
}

func (srv *grpcServer) Poll(ctx context.Context, req *api.ActionPollRequest) (*api.Actions, error) {
	action, err := srv.ActionService.Poll(req.ActionType, int(req.BatchSize))
	if err != nil {
		switch err.(type) {
		case persistence.StorageLayerError:
			return nil, &api.StorageLayerError{}
		}
	}
	return action, nil
}

func (srv *grpcServer) Push(ctx context.Context, req *api.ActionResult) (*api.ActionResultPushResponse, error) {
	err := srv.ActionService.Push(req)
	if err != nil {
		return &api.ActionResultPushResponse{
			Status: false,
		}, err
	}
	return &api.ActionResultPushResponse{Status: true}, nil
}

func (s *grpcServer) GetServers(ctx context.Context, req *api.GetServersRequest) (*api.GetServersResponse, error) {
	s.ClusterRefresher.RefreshCluster()
	servers, err := s.GetServerer.GetServers()
	if err != nil {
		return nil, err
	}
	return &api.GetServersResponse{Servers: servers}, nil
}

func (s *grpcServer) PollStream(req *api.ActionPollRequest, stream api.ActionService_PollStreamServer) error {
	ch, _ := s.ActionService.PollStream(req.ActionType)
	errorCh := make(chan error)
	go func() {
		for {
			act := <-ch
			if err := stream.Send(act); err != nil {
				logger.Error("error sending stream", zap.Error(err))
				errorCh <- err
			}
		}
	}()
	err := <-errorCh
	return err
}
