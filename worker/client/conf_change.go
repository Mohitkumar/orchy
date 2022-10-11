package client

import (
	"context"
	"sort"
	"time"

	api_v1 "github.com/mohitkumar/orchy/api/v1"
)

type clusterConf struct {
	servers  []*api_v1.Server
	stopChan chan struct{}
	clients  []*RpcClient
	client   *RpcClient
}

func newclusterConf(serverUrl string) *clusterConf {
	client, err := NewRpcClient(serverUrl)
	if err != nil {
		panic(err)
	}
	return &clusterConf{
		stopChan: make(chan struct{}),
		client:   client,
	}
}

func (c *clusterConf) add(client *RpcClient) {
	c.clients = append(c.clients, client)
}

func (c *clusterConf) equals(srvs []*api_v1.Server) bool {
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
func (c *clusterConf) start() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			req := &api_v1.GetServersRequest{}
			res, err := c.client.GetApiClient().GetServers(context.Background(), req)
			if err == nil {
				if !c.equals(res.Servers) {
					for _, cl := range c.clients {
						cl.Refresh()
					}
				}
			}
		}
	}
}

func (c *clusterConf) stop() {
	c.stopChan <- struct{}{}
}
