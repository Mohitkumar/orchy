package cluster

import (
	"sync"

	"github.com/buraksezer/consistent"
	api_v1 "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/util"
	"github.com/spaolacci/murmur3"
	"go.uber.org/zap"
)

type hasher struct {
}

func NewHasher() *hasher {
	return &hasher{}
}

type RingConfig struct {
	PartitionCount int
}

func (h hasher) Sum64(data []byte) uint64 {
	return murmur3.Sum64(data)
}

type Ring struct {
	RingConfig
	hring     *consistent.Consistent
	nodes     map[string]Node
	temp      map[string]Node
	localNode Node
	mu        sync.Mutex
}

type Node struct {
	name string
	addr string
}

func (n Node) String() string {
	return n.name
}

func NewRing(c RingConfig) *Ring {
	cfg := consistent.Config{
		PartitionCount:    c.PartitionCount,
		ReplicationFactor: 20,
		Load:              1.25,
		Hasher:            NewHasher(),
	}
	hr := consistent.New(nil, cfg)
	return &Ring{
		RingConfig: c,
		hring:      hr,
		nodes:      make(map[string]Node),
		temp:       make(map[string]Node),
	}
}

func (r *Ring) Join(name, addr string, isLocal bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.nodes[name]; ok {
		return nil
	}
	node := Node{
		name: name,
		addr: addr,
	}
	if isLocal {
		logger.Info("adding member to cluster", zap.String("node", name), zap.String("address", addr))
		r.localNode = node
		r.nodes[name] = node
		r.hring.Add(node)
	} else {
		r.temp[name] = node
	}
	return nil
}

func (r *Ring) Leave(name string) error {
	logger.Info("removing member from cluster", zap.String("node", name))
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.nodes, name)
	delete(r.temp, name)
	r.hring.Remove(name)
	return nil
}

func (r *Ring) GetPartition(key string) int {
	return r.hring.FindPartitionID([]byte(key))
}

func (r *Ring) GetPartitions() []int {
	i := 0
	partitions := make([]int, 0)
	for i < r.PartitionCount {
		owner := r.hring.GetPartitionOwner(i)
		if owner.String() == r.localNode.name {
			partitions = append(partitions, i)
		}
		i++
	}
	util.Shuffle(partitions)
	return partitions
}

func (r *Ring) GetServers() ([]*api_v1.Server, error) {
	nodes := r.nodes
	servers := make([]*api_v1.Server, 0, len(nodes))
	for _, node := range nodes {
		srv := &api_v1.Server{
			Id:      node.name,
			RpcAddr: node.addr,
		}
		servers = append(servers, srv)
	}
	return servers, nil
}

func (r *Ring) RefreshCluster() {
	r.copyNodes()
}

func (r *Ring) copyNodes() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for name, node := range r.temp {
		logger.Info("adding member to cluster", zap.String("node", name), zap.String("address", node.addr))
		r.nodes[name] = node
		r.hring.Add(node)
	}
	for k := range r.temp {
		delete(r.temp, k)
	}
}
