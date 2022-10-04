package client

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	api "github.com/mohitkumar/orchy/api/v1"
	_ "github.com/mohitkumar/orchy/worker/lb"
	"github.com/mohitkumar/orchy/worker/logger"
)

type RpcClient struct {
	serverUrl         string
	conn              *grpc.ClientConn
	taskServiceClient api.TaskServiceClient
}

func NewRpcClient(serverAddress string) (*RpcClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("orchy:///%s", serverAddress), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &RpcClient{
		serverUrl:         serverAddress,
		conn:              conn,
		taskServiceClient: api.NewTaskServiceClient(conn),
	}, nil
}

func (c *RpcClient) Close() error {
	return c.conn.Close()
}

func (c *RpcClient) Refresh() error {
	c.Close()
	conn, err := grpc.Dial(fmt.Sprintf("orchy:///%s", c.serverUrl), grpc.WithInsecure(), grpc.WithReturnConnectionError())
	if err != nil {
		logger.Error("grpc server unavailable", zap.String("server", c.serverUrl))
		return err
	}
	c.conn = conn
	c.taskServiceClient = api.NewTaskServiceClient(conn)
	return nil
}

func (c *RpcClient) GetApiClient() api.TaskServiceClient {
	return c.taskServiceClient
}
