package client

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	api "github.com/mohitkumar/orchy/api/v1"
	"google.golang.org/grpc"
)

type ClusterConf struct {
	servers             []*api.Server
	stopChan            chan struct{}
	clients             []*RpcClient
	actionServiceClient api.ActionServiceClient
	wg                  *sync.WaitGroup
}

func NewclusterConf(serverUrl string, wg *sync.WaitGroup) *ClusterConf {
	conn, err := grpc.Dial(fmt.Sprintf("orchy:///%s", serverUrl), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	actionServiceClient := api.NewActionServiceClient(conn)
	servers, err := actionServiceClient.GetServers(context.Background(), &api.GetServersRequest{})
	if err != nil {
		panic(err)
	}
	return &ClusterConf{
		stopChan:            make(chan struct{}),
		actionServiceClient: actionServiceClient,
		wg:                  wg,
		servers:             servers.Servers,
	}
}

func (c *ClusterConf) add(client *RpcClient) {
	c.clients = append(c.clients, client)
}

func (c *ClusterConf) equals(srvs []*api.Server) bool {
	if len(c.servers) != len(srvs) {
		return false
	}
	sort.Slice(c.servers, func(i, j int) bool {
		return c.servers[i].Id < c.servers[j].Id
	})
	sort.Slice(srvs, func(i, j int) bool {
		return srvs[i].Id < srvs[j].Id
	})
	for i := 0; i < len(srvs); i++ {
		if c.servers[i].Id != srvs[i].Id && c.servers[i].RpcAddr != srvs[i].RpcAddr {
			return false
		}
	}
	return true
}
func (c *ClusterConf) Start() {
	ticker := time.NewTicker(5 * time.Second)
	defer c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.stopChan:
				return
			case <-ticker.C:
				req := &api.GetServersRequest{}
				res, err := c.actionServiceClient.GetServers(context.Background(), req)
				if err == nil {
					if !c.equals(res.Servers) {
						c.servers = res.Servers
						for _, cl := range c.clients {
							cl.Refresh()
						}
					}
				}
			}
		}
	}()
}

func (c *ClusterConf) Stop() {
	c.stopChan <- struct{}{}
}
