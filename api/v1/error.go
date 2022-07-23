package api_v1

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type PollError struct {
	QueueName string
}

func (e PollError) GRPCStatus() *status.Status {
	st := status.New(codes.NotFound, fmt.Sprintf("no task available for execution in queue %s", e.QueueName))
	msg := fmt.Sprintf("no task available for execution in queue %s", e.QueueName)
	d := &errdetails.LocalizedMessage{
		Locale:  "en-US",
		Message: msg,
	}
	std, err := st.WithDetails(d)
	if err != nil {
		return st
	}
	return std
}

func (e PollError) Error() string {
	return e.GRPCStatus().Err().Error()
}

type StorageLayerError struct{}

func (e StorageLayerError) GRPCStatus() *status.Status {
	st := status.New(codes.Internal, fmt.Sprintf("error in underline storage layer"))
	msg := "error in underline storage layer"
	d := &errdetails.LocalizedMessage{
		Locale:  "en-US",
		Message: msg,
	}
	std, err := st.WithDetails(d)
	if err != nil {
		return st
	}
	return std
}

func (e StorageLayerError) Error() string {
	return e.GRPCStatus().Err().Error()
}
