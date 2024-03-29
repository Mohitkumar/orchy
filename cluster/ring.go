package cluster

import (
	"sync"

	"github.com/buraksezer/consistent"
	api_v1 "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/util"
	"github.com/spaolacci/murmur3"
	"go.uber.org/zap"
)

type hasher struct {
}

func NewHasher() *hasher {
	return &hasher{}
}

func (h hasher) Sum64(data []byte) uint64 {
	return murmur3.Sum64(data)
}

type Ring struct {
	partitionCount int
	hring          *consistent.Consistent
	nodes          map[string]Node
	localNode      Node
	rebalancer     func([]int)
	mu             sync.Mutex
}

type Node struct {
	name string
	addr string
}

func (n Node) String() string {
	return n.name
}

func NewRing(paritionCount int) *Ring {
	cfg := consistent.Config{
		PartitionCount:    paritionCount,
		ReplicationFactor: 20,
		Load:              1.25,
		Hasher:            NewHasher(),
	}
	hr := consistent.New(nil, cfg)
	return &Ring{
		partitionCount: paritionCount,
		hring:          hr,
		nodes:          make(map[string]Node),
	}
}

func (r *Ring) SetRebalancer(reb func([]int)) {
	r.rebalancer = reb
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
	logger.Info("adding member to cluster", zap.String("node", name), zap.String("address", addr))
	if isLocal {
		r.localNode = node
	}
	r.nodes[name] = node
	r.hring.Add(node)
	r.rebalancer(r.GetPartitions())
	return nil
}

func (r *Ring) Leave(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	logger.Info("removing member from cluster", zap.String("node", name))
	delete(r.nodes, name)
	r.hring.Remove(name)
	r.rebalancer(r.GetPartitions())
	return nil
}

func (r *Ring) GetPartition(key string) int {
	return r.hring.FindPartitionID([]byte(key))
}

func (r *Ring) GetPartitions() []int {
	i := 0
	partitions := make([]int, 0)
	for i < r.partitionCount {
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
