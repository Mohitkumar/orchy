// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: api/v1/flow.proto

package api_v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type TaskResult_Status int32

const (
	TaskResult_SUCCESS TaskResult_Status = 0
	TaskResult_FAIL    TaskResult_Status = 1
)

// Enum value maps for TaskResult_Status.
var (
	TaskResult_Status_name = map[int32]string{
		0: "SUCCESS",
		1: "FAIL",
	}
	TaskResult_Status_value = map[string]int32{
		"SUCCESS": 0,
		"FAIL":    1,
	}
)

func (x TaskResult_Status) Enum() *TaskResult_Status {
	p := new(TaskResult_Status)
	*p = x
	return p
}

func (x TaskResult_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TaskResult_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_api_v1_flow_proto_enumTypes[0].Descriptor()
}

func (TaskResult_Status) Type() protoreflect.EnumType {
	return &file_api_v1_flow_proto_enumTypes[0]
}

func (x TaskResult_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TaskResult_Status.Descriptor instead.
func (TaskResult_Status) EnumDescriptor() ([]byte, []int) {
	return file_api_v1_flow_proto_rawDescGZIP(), []int{2, 0}
}

type Task struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WorkflowName string                     `protobuf:"bytes,1,opt,name=workflowName,proto3" json:"workflowName,omitempty"`
	FlowId       string                     `protobuf:"bytes,2,opt,name=flowId,proto3" json:"flowId,omitempty"`
	Data         map[string]*structpb.Value `protobuf:"bytes,3,rep,name=data,proto3" json:"data,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	ActionId     int32                      `protobuf:"varint,4,opt,name=actionId,proto3" json:"actionId,omitempty"`
	TaskName     string                     `protobuf:"bytes,5,opt,name=taskName,proto3" json:"taskName,omitempty"`
	RetryCount   int32                      `protobuf:"varint,6,opt,name=retryCount,proto3" json:"retryCount,omitempty"`
}

func (x *Task) Reset() {
	*x = Task{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_flow_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Task) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Task) ProtoMessage() {}

func (x *Task) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_flow_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Task.ProtoReflect.Descriptor instead.
func (*Task) Descriptor() ([]byte, []int) {
	return file_api_v1_flow_proto_rawDescGZIP(), []int{0}
}

func (x *Task) GetWorkflowName() string {
	if x != nil {
		return x.WorkflowName
	}
	return ""
}

func (x *Task) GetFlowId() string {
	if x != nil {
		return x.FlowId
	}
	return ""
}

func (x *Task) GetData() map[string]*structpb.Value {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Task) GetActionId() int32 {
	if x != nil {
		return x.ActionId
	}
	return 0
}

func (x *Task) GetTaskName() string {
	if x != nil {
		return x.TaskName
	}
	return ""
}

func (x *Task) GetRetryCount() int32 {
	if x != nil {
		return x.RetryCount
	}
	return 0
}

type Tasks struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tasks []*Task `protobuf:"bytes,1,rep,name=tasks,proto3" json:"tasks,omitempty"`
}

func (x *Tasks) Reset() {
	*x = Tasks{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_flow_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tasks) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tasks) ProtoMessage() {}

func (x *Tasks) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_flow_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tasks.ProtoReflect.Descriptor instead.
func (*Tasks) Descriptor() ([]byte, []int) {
	return file_api_v1_flow_proto_rawDescGZIP(), []int{1}
}

func (x *Tasks) GetTasks() []*Task {
	if x != nil {
		return x.Tasks
	}
	return nil
}

type TaskResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WorkflowName string                     `protobuf:"bytes,1,opt,name=workflowName,proto3" json:"workflowName,omitempty"`
	TaskName     string                     `protobuf:"bytes,2,opt,name=taskName,proto3" json:"taskName,omitempty"`
	FlowId       string                     `protobuf:"bytes,3,opt,name=flowId,proto3" json:"flowId,omitempty"`
	ActionId     int32                      `protobuf:"varint,4,opt,name=actionId,proto3" json:"actionId,omitempty"`
	Data         map[string]*structpb.Value `protobuf:"bytes,5,rep,name=data,proto3" json:"data,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Status       TaskResult_Status          `protobuf:"varint,6,opt,name=status,proto3,enum=TaskResult_Status" json:"status,omitempty"`
	RetryCount   int32                      `protobuf:"varint,7,opt,name=retryCount,proto3" json:"retryCount,omitempty"`
}

func (x *TaskResult) Reset() {
	*x = TaskResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_flow_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskResult) ProtoMessage() {}

func (x *TaskResult) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_flow_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskResult.ProtoReflect.Descriptor instead.
func (*TaskResult) Descriptor() ([]byte, []int) {
	return file_api_v1_flow_proto_rawDescGZIP(), []int{2}
}

func (x *TaskResult) GetWorkflowName() string {
	if x != nil {
		return x.WorkflowName
	}
	return ""
}

func (x *TaskResult) GetTaskName() string {
	if x != nil {
		return x.TaskName
	}
	return ""
}

func (x *TaskResult) GetFlowId() string {
	if x != nil {
		return x.FlowId
	}
	return ""
}

func (x *TaskResult) GetActionId() int32 {
	if x != nil {
		return x.ActionId
	}
	return 0
}

func (x *TaskResult) GetData() map[string]*structpb.Value {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *TaskResult) GetStatus() TaskResult_Status {
	if x != nil {
		return x.Status
	}
	return TaskResult_SUCCESS
}

func (x *TaskResult) GetRetryCount() int32 {
	if x != nil {
		return x.RetryCount
	}
	return 0
}

type TaskDef struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name              string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	RetryCount        int32  `protobuf:"varint,2,opt,name=retryCount,proto3" json:"retryCount,omitempty"`
	RetryAfterSeconds int32  `protobuf:"varint,3,opt,name=retryAfterSeconds,proto3" json:"retryAfterSeconds,omitempty"`
	RetryPolicy       string `protobuf:"bytes,4,opt,name=retryPolicy,proto3" json:"retryPolicy,omitempty"`
	TimeoutSeconds    int32  `protobuf:"varint,5,opt,name=timeoutSeconds,proto3" json:"timeoutSeconds,omitempty"`
}

func (x *TaskDef) Reset() {
	*x = TaskDef{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_flow_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskDef) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskDef) ProtoMessage() {}

func (x *TaskDef) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_flow_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskDef.ProtoReflect.Descriptor instead.
func (*TaskDef) Descriptor() ([]byte, []int) {
	return file_api_v1_flow_proto_rawDescGZIP(), []int{3}
}

func (x *TaskDef) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *TaskDef) GetRetryCount() int32 {
	if x != nil {
		return x.RetryCount
	}
	return 0
}

func (x *TaskDef) GetRetryAfterSeconds() int32 {
	if x != nil {
		return x.RetryAfterSeconds
	}
	return 0
}

func (x *TaskDef) GetRetryPolicy() string {
	if x != nil {
		return x.RetryPolicy
	}
	return ""
}

func (x *TaskDef) GetTimeoutSeconds() int32 {
	if x != nil {
		return x.TimeoutSeconds
	}
	return 0
}

type TaskDefSaveResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status bool `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *TaskDefSaveResponse) Reset() {
	*x = TaskDefSaveResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_flow_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskDefSaveResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskDefSaveResponse) ProtoMessage() {}

func (x *TaskDefSaveResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_flow_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskDefSaveResponse.ProtoReflect.Descriptor instead.
func (*TaskDefSaveResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_flow_proto_rawDescGZIP(), []int{4}
}

func (x *TaskDefSaveResponse) GetStatus() bool {
	if x != nil {
		return x.Status
	}
	return false
}

type TaskPollRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskType  string `protobuf:"bytes,1,opt,name=taskType,proto3" json:"taskType,omitempty"`
	BatchSize int32  `protobuf:"varint,2,opt,name=batchSize,proto3" json:"batchSize,omitempty"`
}

func (x *TaskPollRequest) Reset() {
	*x = TaskPollRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_flow_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskPollRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskPollRequest) ProtoMessage() {}

func (x *TaskPollRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_flow_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskPollRequest.ProtoReflect.Descriptor instead.
func (*TaskPollRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_flow_proto_rawDescGZIP(), []int{5}
}

func (x *TaskPollRequest) GetTaskType() string {
	if x != nil {
		return x.TaskType
	}
	return ""
}

func (x *TaskPollRequest) GetBatchSize() int32 {
	if x != nil {
		return x.BatchSize
	}
	return 0
}

type TaskResultPushResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status bool `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *TaskResultPushResponse) Reset() {
	*x = TaskResultPushResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_flow_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskResultPushResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskResultPushResponse) ProtoMessage() {}

func (x *TaskResultPushResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_flow_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskResultPushResponse.ProtoReflect.Descriptor instead.
func (*TaskResultPushResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_flow_proto_rawDescGZIP(), []int{6}
}

func (x *TaskResultPushResponse) GetStatus() bool {
	if x != nil {
		return x.Status
	}
	return false
}

type Server struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	RpcAddr string `protobuf:"bytes,2,opt,name=rpc_addr,json=rpcAddr,proto3" json:"rpc_addr,omitempty"`
}

func (x *Server) Reset() {
	*x = Server{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_flow_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Server) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Server) ProtoMessage() {}

func (x *Server) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_flow_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Server.ProtoReflect.Descriptor instead.
func (*Server) Descriptor() ([]byte, []int) {
	return file_api_v1_flow_proto_rawDescGZIP(), []int{7}
}

func (x *Server) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Server) GetRpcAddr() string {
	if x != nil {
		return x.RpcAddr
	}
	return ""
}

type GetServersRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetServersRequest) Reset() {
	*x = GetServersRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_flow_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetServersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetServersRequest) ProtoMessage() {}

func (x *GetServersRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_flow_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetServersRequest.ProtoReflect.Descriptor instead.
func (*GetServersRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_flow_proto_rawDescGZIP(), []int{8}
}

type GetServersResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Servers []*Server `protobuf:"bytes,1,rep,name=servers,proto3" json:"servers,omitempty"`
}

func (x *GetServersResponse) Reset() {
	*x = GetServersResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_flow_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetServersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetServersResponse) ProtoMessage() {}

func (x *GetServersResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_flow_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetServersResponse.ProtoReflect.Descriptor instead.
func (*GetServersResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_flow_proto_rawDescGZIP(), []int{9}
}

func (x *GetServersResponse) GetServers() []*Server {
	if x != nil {
		return x.Servers
	}
	return nil
}

var File_api_v1_flow_proto protoreflect.FileDescriptor

var file_api_v1_flow_proto_rawDesc = []byte{
	0x0a, 0x11, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x90, 0x02, 0x0a, 0x04, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x22, 0x0a, 0x0c, 0x77, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x66, 0x6c, 0x6f, 0x77, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x66, 0x6c, 0x6f, 0x77, 0x49, 0x64, 0x12, 0x23, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x2e, 0x44, 0x61, 0x74, 0x61,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1a, 0x0a, 0x08, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x74, 0x61, 0x73, 0x6b, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x61, 0x73, 0x6b, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x72, 0x65, 0x74, 0x72, 0x79, 0x43, 0x6f, 0x75, 0x6e,
	0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x72, 0x65, 0x74, 0x72, 0x79, 0x43, 0x6f,
	0x75, 0x6e, 0x74, 0x1a, 0x4f, 0x0a, 0x09, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x22, 0x24, 0x0a, 0x05, 0x54, 0x61, 0x73, 0x6b, 0x73, 0x12, 0x1b, 0x0a,
	0x05, 0x74, 0x61, 0x73, 0x6b, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x05, 0x2e, 0x54,
	0x61, 0x73, 0x6b, 0x52, 0x05, 0x74, 0x61, 0x73, 0x6b, 0x73, 0x22, 0xe9, 0x02, 0x0a, 0x0a, 0x54,
	0x61, 0x73, 0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x22, 0x0a, 0x0c, 0x77, 0x6f, 0x72,
	0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0c, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a,
	0x08, 0x74, 0x61, 0x73, 0x6b, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x74, 0x61, 0x73, 0x6b, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x6c, 0x6f,
	0x77, 0x49, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x66, 0x6c, 0x6f, 0x77, 0x49,
	0x64, 0x12, 0x1a, 0x0a, 0x08, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x08, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x29, 0x0a,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x54, 0x61,
	0x73, 0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x2a, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x12, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x52,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x72, 0x65, 0x74, 0x72, 0x79, 0x43, 0x6f, 0x75,
	0x6e, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x72, 0x65, 0x74, 0x72, 0x79, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x1a, 0x4f, 0x0a, 0x09, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x1f, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x0b, 0x0a, 0x07, 0x53, 0x55, 0x43, 0x43, 0x45, 0x53, 0x53, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04,
	0x46, 0x41, 0x49, 0x4c, 0x10, 0x01, 0x22, 0xb5, 0x01, 0x0a, 0x07, 0x54, 0x61, 0x73, 0x6b, 0x44,
	0x65, 0x66, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x72, 0x65, 0x74, 0x72, 0x79, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x72, 0x65, 0x74, 0x72,
	0x79, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x2c, 0x0a, 0x11, 0x72, 0x65, 0x74, 0x72, 0x79, 0x41,
	0x66, 0x74, 0x65, 0x72, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x11, 0x72, 0x65, 0x74, 0x72, 0x79, 0x41, 0x66, 0x74, 0x65, 0x72, 0x53, 0x65, 0x63,
	0x6f, 0x6e, 0x64, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x72, 0x65, 0x74, 0x72, 0x79, 0x50, 0x6f, 0x6c,
	0x69, 0x63, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x72, 0x65, 0x74, 0x72, 0x79,
	0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x26, 0x0a, 0x0e, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75,
	0x74, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e,
	0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x22, 0x2d,
	0x0a, 0x13, 0x54, 0x61, 0x73, 0x6b, 0x44, 0x65, 0x66, 0x53, 0x61, 0x76, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x4b, 0x0a,
	0x0f, 0x54, 0x61, 0x73, 0x6b, 0x50, 0x6f, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x1a, 0x0a, 0x08, 0x74, 0x61, 0x73, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x74, 0x61, 0x73, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1c, 0x0a, 0x09,
	0x62, 0x61, 0x74, 0x63, 0x68, 0x53, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x09, 0x62, 0x61, 0x74, 0x63, 0x68, 0x53, 0x69, 0x7a, 0x65, 0x22, 0x30, 0x0a, 0x16, 0x54, 0x61,
	0x73, 0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x50, 0x75, 0x73, 0x68, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x33, 0x0a, 0x06,
	0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x72, 0x70, 0x63, 0x5f, 0x61, 0x64,
	0x64, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x72, 0x70, 0x63, 0x41, 0x64, 0x64,
	0x72, 0x22, 0x13, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x37, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x53, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x21, 0x0a, 0x07,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x07, 0x2e,
	0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x32,
	0xcb, 0x01, 0x0a, 0x0b, 0x54, 0x61, 0x73, 0x6b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x2f, 0x0a, 0x0b, 0x53, 0x61, 0x76, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x44, 0x65, 0x66, 0x12, 0x08,
	0x2e, 0x54, 0x61, 0x73, 0x6b, 0x44, 0x65, 0x66, 0x1a, 0x14, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x44,
	0x65, 0x66, 0x53, 0x61, 0x76, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x12, 0x22, 0x0a, 0x04, 0x50, 0x6f, 0x6c, 0x6c, 0x12, 0x10, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x50,
	0x6f, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x06, 0x2e, 0x54, 0x61, 0x73,
	0x6b, 0x73, 0x22, 0x00, 0x12, 0x2e, 0x0a, 0x04, 0x50, 0x75, 0x73, 0x68, 0x12, 0x0b, 0x2e, 0x54,
	0x61, 0x73, 0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x1a, 0x17, 0x2e, 0x54, 0x61, 0x73, 0x6b,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x50, 0x75, 0x73, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x37, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x53, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x73, 0x12, 0x12, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x1e, 0x5a,
	0x1c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x6f, 0x68, 0x69,
	0x74, 0x6b, 0x75, 0x6d, 0x61, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x5f, 0x76, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_v1_flow_proto_rawDescOnce sync.Once
	file_api_v1_flow_proto_rawDescData = file_api_v1_flow_proto_rawDesc
)

func file_api_v1_flow_proto_rawDescGZIP() []byte {
	file_api_v1_flow_proto_rawDescOnce.Do(func() {
		file_api_v1_flow_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_v1_flow_proto_rawDescData)
	})
	return file_api_v1_flow_proto_rawDescData
}

var file_api_v1_flow_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_api_v1_flow_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_api_v1_flow_proto_goTypes = []interface{}{
	(TaskResult_Status)(0),         // 0: TaskResult.Status
	(*Task)(nil),                   // 1: Task
	(*Tasks)(nil),                  // 2: Tasks
	(*TaskResult)(nil),             // 3: TaskResult
	(*TaskDef)(nil),                // 4: TaskDef
	(*TaskDefSaveResponse)(nil),    // 5: TaskDefSaveResponse
	(*TaskPollRequest)(nil),        // 6: TaskPollRequest
	(*TaskResultPushResponse)(nil), // 7: TaskResultPushResponse
	(*Server)(nil),                 // 8: Server
	(*GetServersRequest)(nil),      // 9: GetServersRequest
	(*GetServersResponse)(nil),     // 10: GetServersResponse
	nil,                            // 11: Task.DataEntry
	nil,                            // 12: TaskResult.DataEntry
	(*structpb.Value)(nil),         // 13: google.protobuf.Value
}
var file_api_v1_flow_proto_depIdxs = []int32{
	11, // 0: Task.data:type_name -> Task.DataEntry
	1,  // 1: Tasks.tasks:type_name -> Task
	12, // 2: TaskResult.data:type_name -> TaskResult.DataEntry
	0,  // 3: TaskResult.status:type_name -> TaskResult.Status
	8,  // 4: GetServersResponse.servers:type_name -> Server
	13, // 5: Task.DataEntry.value:type_name -> google.protobuf.Value
	13, // 6: TaskResult.DataEntry.value:type_name -> google.protobuf.Value
	4,  // 7: TaskService.SaveTaskDef:input_type -> TaskDef
	6,  // 8: TaskService.Poll:input_type -> TaskPollRequest
	3,  // 9: TaskService.Push:input_type -> TaskResult
	9,  // 10: TaskService.GetServers:input_type -> GetServersRequest
	5,  // 11: TaskService.SaveTaskDef:output_type -> TaskDefSaveResponse
	2,  // 12: TaskService.Poll:output_type -> Tasks
	7,  // 13: TaskService.Push:output_type -> TaskResultPushResponse
	10, // 14: TaskService.GetServers:output_type -> GetServersResponse
	11, // [11:15] is the sub-list for method output_type
	7,  // [7:11] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_api_v1_flow_proto_init() }
func file_api_v1_flow_proto_init() {
	if File_api_v1_flow_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_v1_flow_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Task); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_v1_flow_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tasks); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_v1_flow_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskResult); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_v1_flow_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskDef); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_v1_flow_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskDefSaveResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_v1_flow_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskPollRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_v1_flow_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskResultPushResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_v1_flow_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Server); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_v1_flow_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetServersRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_v1_flow_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetServersResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_v1_flow_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_v1_flow_proto_goTypes,
		DependencyIndexes: file_api_v1_flow_proto_depIdxs,
		EnumInfos:         file_api_v1_flow_proto_enumTypes,
		MessageInfos:      file_api_v1_flow_proto_msgTypes,
	}.Build()
	File_api_v1_flow_proto = out.File
	file_api_v1_flow_proto_rawDesc = nil
	file_api_v1_flow_proto_goTypes = nil
	file_api_v1_flow_proto_depIdxs = nil
}
