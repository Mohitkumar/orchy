package agent

import (
	"fmt"
	"net"
	"sync"

	"github.com/mohitkumar/orchy/server/cluster"
	"github.com/mohitkumar/orchy/server/cluster/executor"
	"github.com/mohitkumar/orchy/server/config"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/flow"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/rest"
	"github.com/mohitkumar/orchy/server/rpc"
	"github.com/mohitkumar/orchy/server/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Agent struct {
	Config                   config.Config
	ring                     *cluster.Ring
	membership               *cluster.Membership
	diContainer              *container.DIContiner
	flowService              *flow.FlowService
	httpServer               *rest.Server
	grpcServer               *grpc.Server
	actionExecutionService   *service.ActionExecutionService
	workflowExecutionService *service.WorkflowExecutionService
	executors                *executor.Executors
	shutdown                 bool
	shutdowns                chan struct{}
	shutdownLock             sync.Mutex
	wg                       sync.WaitGroup
}

func New(config config.Config) (*Agent, error) {
	a := &Agent{
		Config:    config,
		shutdowns: make(chan struct{}),
	}
	setup := []func() error{
		a.setupCluster,
		a.setupDiContainer,
		a.setupFlowService,
		a.setupWorkflowExecutionService,
		a.setupActionExecutorService,
		a.setupExecutors,
		a.setupHttpServer,
		a.setupGrpcServer,
	}
	for _, fn := range setup {
		if err := fn(); err != nil {
			return nil, err
		}
	}
	return a, nil
}

func (a *Agent) setupCluster() error {
	r := cluster.NewRing(a.Config.RingConfig)
	a.ring = r
	mem, err := cluster.New(r, a.Config.ClusterConfig)
	if err != nil {
		return err
	}
	a.membership = mem
	return nil
}

func (a *Agent) setupDiContainer() error {
	a.diContainer = container.NewDiContainer(a.ring)
	a.diContainer.Init(a.Config)
	return nil
}

func (a *Agent) setupFlowService() error {
	a.flowService = flow.NewFlowService(a.diContainer, &a.wg)
	a.flowService.Start()
	return nil
}

func (a *Agent) setupWorkflowExecutionService() error {
	a.workflowExecutionService = service.NewWorkflowExecutionService(a.flowService)
	return nil
}

func (a *Agent) setupActionExecutorService() error {
	a.actionExecutionService = service.NewActionExecutionService(a.diContainer, a.flowService)
	return nil
}

func (a *Agent) setupExecutors() error {
	a.executors = executor.NewExecutors(a.diContainer.GetShards())
	a.executors.InitExecutors(a.Config.RingConfig.PartitionCount, a.diContainer, a.flowService, &a.wg)
	a.executors.StartAll()
	return nil
}
func (a *Agent) setupHttpServer() error {
	var err error
	a.httpServer, err = rest.NewServer(a.Config.HttpPort, a.diContainer, a.workflowExecutionService)
	if err != nil {
		return err
	}
	return nil
}

func (a *Agent) setupGrpcServer() error {
	var err error
	conf := &rpc.GrpcConfig{
		ActionService:           a.actionExecutionService,
		ActionDefinitionService: a.diContainer.GetMetadataStorage(),
		GetServerer:             a.ring,
		ClusterRefresher:        a.membership,
	}
	a.grpcServer, err = rpc.NewGrpcServer(conf)
	if err != nil {
		return err
	}
	return nil
}

func (a *Agent) Start() error {
	var err error
	go func() error {
		err = a.httpServer.Start()
		if err != nil {
			_ = a.Shutdown()
			panic(err)
		}
		return nil
	}()

	go func() error {
		logger.Info("startting grpc server on", zap.Int("port", a.Config.GrpcPort))
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.Config.GrpcPort))
		if err != nil {
			panic(err)
		}

		if err := a.grpcServer.Serve(lis); err != nil {
			_ = a.Shutdown()
			panic(err)
		}
		return nil
	}()
	return nil
}

func (a *Agent) Shutdown() error {
	logger.Info("shutting down server")
	a.shutdownLock.Lock()
	defer a.shutdownLock.Unlock()
	if a.shutdown {
		return nil
	}
	a.shutdown = true
	close(a.shutdowns)

	shutdown := []func() error{
		a.executors.StopAll,
		a.flowService.Stop,
		a.httpServer.Stop,
		func() error {
			logger.Info("stopping grpc server")
			a.grpcServer.Stop()
			return nil
		},
	}
	for _, fn := range shutdown {
		if err := fn(); err != nil {
			return err
		}
	}
	logger.Info("waiting for all services to shutdown...")
	a.wg.Wait()
	return nil
}
