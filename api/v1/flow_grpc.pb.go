// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: api/v1/flow.proto

package api_v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TaskServiceClient is the client API for TaskService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TaskServiceClient interface {
	SaveTaskDef(ctx context.Context, in *TaskDef, opts ...grpc.CallOption) (*TaskDefSaveResponse, error)
	Poll(ctx context.Context, in *TaskPollRequest, opts ...grpc.CallOption) (*Task, error)
	Push(ctx context.Context, in *TaskResult, opts ...grpc.CallOption) (*TaskResultPushResponse, error)
}

type taskServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTaskServiceClient(cc grpc.ClientConnInterface) TaskServiceClient {
	return &taskServiceClient{cc}
}

func (c *taskServiceClient) SaveTaskDef(ctx context.Context, in *TaskDef, opts ...grpc.CallOption) (*TaskDefSaveResponse, error) {
	out := new(TaskDefSaveResponse)
	err := c.cc.Invoke(ctx, "/TaskService/SaveTaskDef", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskServiceClient) Poll(ctx context.Context, in *TaskPollRequest, opts ...grpc.CallOption) (*Task, error) {
	out := new(Task)
	err := c.cc.Invoke(ctx, "/TaskService/Poll", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskServiceClient) Push(ctx context.Context, in *TaskResult, opts ...grpc.CallOption) (*TaskResultPushResponse, error) {
	out := new(TaskResultPushResponse)
	err := c.cc.Invoke(ctx, "/TaskService/Push", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TaskServiceServer is the server API for TaskService service.
// All implementations must embed UnimplementedTaskServiceServer
// for forward compatibility
type TaskServiceServer interface {
	SaveTaskDef(context.Context, *TaskDef) (*TaskDefSaveResponse, error)
	Poll(context.Context, *TaskPollRequest) (*Task, error)
	Push(context.Context, *TaskResult) (*TaskResultPushResponse, error)
	mustEmbedUnimplementedTaskServiceServer()
}

// UnimplementedTaskServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTaskServiceServer struct {
}

func (UnimplementedTaskServiceServer) SaveTaskDef(context.Context, *TaskDef) (*TaskDefSaveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SaveTaskDef not implemented")
}
func (UnimplementedTaskServiceServer) Poll(context.Context, *TaskPollRequest) (*Task, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Poll not implemented")
}
func (UnimplementedTaskServiceServer) Push(context.Context, *TaskResult) (*TaskResultPushResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Push not implemented")
}
func (UnimplementedTaskServiceServer) mustEmbedUnimplementedTaskServiceServer() {}

// UnsafeTaskServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TaskServiceServer will
// result in compilation errors.
type UnsafeTaskServiceServer interface {
	mustEmbedUnimplementedTaskServiceServer()
}

func RegisterTaskServiceServer(s grpc.ServiceRegistrar, srv TaskServiceServer) {
	s.RegisterService(&TaskService_ServiceDesc, srv)
}

func _TaskService_SaveTaskDef_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskDef)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).SaveTaskDef(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/TaskService/SaveTaskDef",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).SaveTaskDef(ctx, req.(*TaskDef))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskService_Poll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskPollRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).Poll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/TaskService/Poll",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).Poll(ctx, req.(*TaskPollRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaskService_Push_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskResult)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).Push(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/TaskService/Push",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).Push(ctx, req.(*TaskResult))
	}
	return interceptor(ctx, in, info, handler)
}

// TaskService_ServiceDesc is the grpc.ServiceDesc for TaskService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TaskService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "TaskService",
	HandlerType: (*TaskServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SaveTaskDef",
			Handler:    _TaskService_SaveTaskDef_Handler,
		},
		{
			MethodName: "Poll",
			Handler:    _TaskService_Poll_Handler,
		},
		{
			MethodName: "Push",
			Handler:    _TaskService_Push_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/v1/flow.proto",
}
