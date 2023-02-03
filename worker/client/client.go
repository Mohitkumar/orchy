package client

import (
	"fmt"
	"sync"

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
	mu                  sync.Mutex
}

func NewRpcClient(serverAddress string, clusterConf *ClusterConf) (*RpcClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("orchy:///%s", serverAddress), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := &RpcClient{
		serverUrl:           serverAddress,
		conn:                conn,
		actionServiceClient: api.NewActionServiceClient(conn),
	}
	clusterConf.add(client)
	return client, nil
}

func (c *RpcClient) Close() error {
	return c.conn.Close()
}

func (c *RpcClient) Refresh() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	oldConn := c.conn
	conn, err := grpc.Dial(fmt.Sprintf("orchy:///%s", c.serverUrl), grpc.WithInsecure())
	if err != nil {
		logger.Error("grpc server unavailable", zap.String("server", c.serverUrl))
		return err
	}
	c.conn = conn
	c.actionServiceClient = api.NewActionServiceClient(conn)
	oldConn.Close()
	return nil
}

func (c *RpcClient) GetApiClient() api.ActionServiceClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.actionServiceClient
}
