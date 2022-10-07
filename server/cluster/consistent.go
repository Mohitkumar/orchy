package cluster

import (
	"hash"
	"sync"

	"github.com/buraksezer/consistent"
	api_v1 "github.com/mohitkumar/orchy/api/v1"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/spaolacci/murmur3"
	"go.uber.org/zap"
)

type hasher struct {
	hf hash.Hash64
}

func NewHasher() *hasher {
	return &hasher{
		hf: murmur3.New64(),
	}
}

type RingConfig struct {
	PartitionCount int
}

func (h hasher) Sum64(data []byte) uint64 {
	// you should use a proper hash function for uniformity.
	h.hf.Write(data)
	out := h.hf.Sum64()
	h.hf.Reset()
	return out
}

type Ring struct {
	RingConfig
	hring     *consistent.Consistent
	nodes     map[string]Node
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
	}
}

func (r *Ring) Join(name, addr string, isLocal bool) error {
	logger.Info("adding member to cluster", zap.String("node", name))
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.nodes[name]; ok {
		return nil
	}
	node := Node{
		name: name,
		addr: addr,
	}
	r.nodes[name] = node
	r.hring.Add(node)
	if isLocal {
		r.localNode = node
	}
	return nil
}

func (r *Ring) Leave(name string) error {
	logger.Info("removing member from cluster", zap.String("node", name))
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.nodes, name)
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
