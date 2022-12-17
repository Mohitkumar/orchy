package client

import (
	"fmt"

	"google.golang.org/grpc"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/worker/lb"
)

type RpcClient struct {
	serverUrls        []string
	conn              *grpc.ClientConn
	taskServiceClient api.TaskServiceClient
}

func NewRpcClient(serverAddresses []string) (*RpcClient, error) {
	resolver := lb.NewResolver(serverAddresses)
	resolver.Init()
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", "orchy", serverAddresses[0]), grpc.WithInsecure(), grpc.WithResolvers(resolver))
	if err != nil {
		return nil, err
	}
	return &RpcClient{
		serverUrls:        serverAddresses,
		conn:              conn,
		taskServiceClient: api.NewTaskServiceClient(conn),
	}, nil
}

func (c *RpcClient) Close() error {
	return c.conn.Close()
}

func (c *RpcClient) GetApiClient() api.TaskServiceClient {
	return c.taskServiceClient
}
