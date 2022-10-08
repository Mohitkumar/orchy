package lb

import (
	"strings"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

var _ base.PickerBuilder = (*Picker)(nil)

type Picker struct {
	mu            sync.RWMutex
	servers       []balancer.SubConn
	workerCounter sync.Map
	current       uint64
	logger        *zap.Logger
}

func (p *Picker) Build(buildInfo base.PickerBuildInfo) balancer.Picker {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.logger = zap.L().Named("picker")
	var servers []balancer.SubConn
	for sc := range buildInfo.ReadySCs {
		servers = append(servers, sc)
	}
	p.servers = servers
	return p
}

var _ balancer.Picker = (*Picker)(nil)

func (p *Picker) Pick(info balancer.PickInfo) (
	balancer.PickResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var result balancer.PickResult
	if strings.Contains(info.FullMethodName, "Poll") {
		worker := info.Ctx.Value("worker")
		result.SubConn = p.nextServerPoll(worker.(string))
	} else {
		result.SubConn = p.nextServer()
	}
	if result.SubConn == nil {
		return result, balancer.ErrNoSubConnAvailable
	}
	return result, nil
}

func (p *Picker) nextServerPoll(worker string) balancer.SubConn {
	currentObj, ok := p.workerCounter.Load(worker)
	var cur uint64
	if ok {
		current := currentObj.(uint64)
		cur = atomic.AddUint64(&current, uint64(1))
		p.workerCounter.Store(worker, cur)
	} else {
		p.workerCounter.Store(worker, uint64(0))
		cur = 0
	}
	len := uint64(len(p.servers))
	if len == 0 {
		return nil
	}
	idx := int(cur % len)
	return p.servers[idx]
}
func (p *Picker) nextServer() balancer.SubConn {
	cur := atomic.AddUint64(&p.current, uint64(1))
	len := uint64(len(p.servers))
	if len == 0 {
		return nil
	}
	idx := int(cur % len)
	return p.servers[idx]
}

func init() {
	balancer.Register(
		base.NewBalancerBuilder(Name, &Picker{}, base.Config{}),
	)
}
