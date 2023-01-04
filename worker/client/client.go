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
	serverUrl           string
	conn                *grpc.ClientConn
	actionServiceClient api.ActionServiceClient
}

func NewRpcClient(serverAddress string) (*RpcClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("orchy:///%s", serverAddress), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &RpcClient{
		serverUrl:           serverAddress,
		conn:                conn,
		actionServiceClient: api.NewActionServiceClient(conn),
	}, nil
}

func (c *RpcClient) Close() error {
	return c.conn.Close()
}

func (c *RpcClient) Refresh() error {
	fmt.Println("refresh")
	oldConn := c.conn
	conn, err := grpc.Dial(fmt.Sprintf("orchy:///%s", c.serverUrl), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.Error("grpc server unavailable", zap.String("server", c.serverUrl))
		return err
	}
	c.conn = conn
	c.actionServiceClient = api.NewActionServiceClient(conn)
	oldConn.Close()
	fmt.Println("refresh success")
	return nil
}

func (c *RpcClient) GetApiClient() api.ActionServiceClient {
	return c.actionServiceClient
}
