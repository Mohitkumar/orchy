package worker

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"

	api "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/worker/logger"
)

type client struct {
	serverUrl         string
	conn              *grpc.ClientConn
	taskServiceClient api.TaskServiceClient
}

func newClient(serverAddress string) (*client, error) {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &client{
		serverUrl:         serverAddress,
		conn:              conn,
		taskServiceClient: api.NewTaskServiceClient(conn),
	}, nil
}

func (c *client) Close() error {
	return c.conn.Close()
}

func (c *client) Refresh() error {
	c.Close()
	conn, err := grpc.Dial(c.serverUrl, grpc.WithInsecure(), grpc.WithReturnConnectionError())
	if err != nil {
		logger.Error("grpc server unavailable", zap.String("server", c.serverUrl))
		return err
	}
	c.conn = conn
	c.taskServiceClient = api.NewTaskServiceClient(conn)
	return nil
}

func (c *client) GetApiClient() api.TaskServiceClient {
	return c.taskServiceClient
}
