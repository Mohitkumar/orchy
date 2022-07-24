package rpc

import (
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/model"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

type TaskService interface {
	Poll(taskName string) (*api.Task, error)
	Push(*api.TaskResult) error
}

type TaskDefService interface {
	SaveTask(task model.TaskDef) error
}
type GrpcConfig struct {
	TaskService    TaskService
	TaskDefService TaskDefService
}

type grpcServer struct {
	api.UnimplementedTaskServiceServer
	*GrpcConfig
}

func NewGrpcServer(config *GrpcConfig) (*grpc.Server, error) {
	logger := zap.L().Named("server")
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDurationField(
			func(duration time.Duration) zapcore.Field {
				return zap.Int64(
					"grpc.time_ns",
					duration.Nanoseconds(),
				)
			},
		),
	}
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	err := view.Register(ocgrpc.DefaultServerViews...)
	if err != nil {
		return nil, err
	}
	grpcOpts := make([]grpc.ServerOption, 0)
	grpcOpts = append(grpcOpts,
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				grpc_zap.StreamServerInterceptor(logger, zapOpts...),
			)), grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
		)),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)

	gsrv := grpc.NewServer(grpcOpts...)
	srv := &grpcServer{
		GrpcConfig: config,
	}
	if err != nil {
		return nil, err
	}
	api.RegisterTaskServiceServer(gsrv, srv)
	return gsrv, nil
}
