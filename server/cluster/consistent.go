package cluster

import (
	"hash"

	"github.com/buraksezer/consistent"
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
	hring *consistent.Consistent
}

type Member string

func (m Member) String() string {
	return string(m)
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
	}
}

func (r *Ring) Join(name, addr string) error {
	logger.Info("adding member to cluster", zap.String("node", name))
	r.hring.Add(Member(name))
	return nil
}

func (r *Ring) Leave(name string) error {
	logger.Info("removing member from cluster", zap.String("node", name))
	r.hring.Remove(name)
	return nil
}
