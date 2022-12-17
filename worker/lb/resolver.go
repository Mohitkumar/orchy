package lb

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

type Resolver struct {
	mu            sync.Mutex
	clientConn    resolver.ClientConn
	resolverConn  *grpc.ClientConn
	serviceConfig *serviceconfig.ParseResult
	logger        *zap.Logger
	addrs         []string
}

var _ resolver.Builder = (*Resolver)(nil)

func (r *Resolver) Build(
	target resolver.Target,
	cc resolver.ClientConn,
	opts resolver.BuildOptions,
) (resolver.Resolver, error) {
	r.logger = zap.L().Named("resolver")
	r.clientConn = cc
	var dialOpts []grpc.DialOption
	if opts.DialCreds != nil {
		dialOpts = append(
			dialOpts,
			grpc.WithTransportCredentials(opts.DialCreds),
			grpc.WithBlock(),
		)
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure(), grpc.WithBlock())
	}
	r.serviceConfig = r.clientConn.ParseServiceConfig(
		fmt.Sprintf(`{"loadBalancingConfig":[{"%s":{}}]}`, Name),
	)
	var err error
	r.resolverConn, err = grpc.Dial(target.Endpoint, dialOpts...)
	if err != nil {
		return nil, err
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

const Name = "orchy"

func (r *Resolver) Scheme() string {
	return Name
}

func NewResolver(addrs []string) *Resolver {
	return &Resolver{
		addrs: addrs,
	}
}

func (r *Resolver) Init() {
	resolver.Register(&Resolver{})
}

var _ resolver.Resolver = (*Resolver)(nil)

func (r *Resolver) ResolveNow(opt resolver.ResolveNowOptions) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var addrs []resolver.Address
	for _, addr := range r.addrs {
		addrs = append(addrs, resolver.Address{
			Addr: addr,
		})
	}
	r.clientConn.UpdateState(resolver.State{
		Addresses:     addrs,
		ServiceConfig: r.serviceConfig,
	})
}

func (r *Resolver) Close() {
	if err := r.resolverConn.Close(); err != nil {
		r.logger.Error(
			"failed to close conn",
			zap.Error(err),
		)
	}
}
