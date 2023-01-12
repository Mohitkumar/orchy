package agent

import (
	"fmt"
	"net"
	"sync"

	"github.com/mohitkumar/orchy/server/cluster"
	"github.com/mohitkumar/orchy/server/config"
	"github.com/mohitkumar/orchy/server/engine"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/metadata"
	"github.com/mohitkumar/orchy/server/model"
	rd "github.com/mohitkumar/orchy/server/persistence/redis"
	"github.com/mohitkumar/orchy/server/rest"
	"github.com/mohitkumar/orchy/server/rpc"
	"github.com/mohitkumar/orchy/server/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	v8 "rogchap.com/v8go"
)

type Agent struct {
	Config                   config.Config
	executionChannel         chan model.FlowExecutionRequest
	cluster                  *cluster.Cluster
	metadataService          metadata.MetadataService
	jsvm                     *v8.Isolate
	flowEngine               *engine.FlowEngine
	httpServer               *rest.Server
	grpcServer               *grpc.Server
	actionExecutionService   *service.ActionExecutionService
	workflowExecutionService *service.WorkflowExecutionService
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
		a.setupExecutionChannel,
		a.setupJsVm,
		a.setupMetadataService,
		a.setupCluster,
		a.setupFlowEngine,
		a.setupWorkflowExecutionService,
		a.setupActionExecutorService,
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

func (a *Agent) setupJsVm() error {
	a.jsvm = v8.NewIsolate()
	return nil
}

func (a *Agent) setupMetadataService() error {
	var metadataStorage metadata.MetadataStorage
	switch a.Config.StorageType {
	case config.STORAGE_TYPE_REDIS:
		rdConf := rd.Config{
			Addrs:     a.Config.RedisConfig.Addrs,
			Namespace: a.Config.RedisConfig.Namespace,
		}
		metadataStorage = rd.NewRedisMetadataStorage(rdConf)
	case config.STORAGE_TYPE_INMEM:
	}
	a.metadataService = metadata.NewMetadataService(metadataStorage, a.jsvm)
	return nil
}

func (a *Agent) setupExecutionChannel() error {
	a.executionChannel = make(chan model.FlowExecutionRequest)
	return nil
}
func (a *Agent) setupCluster() error {
	a.cluster = cluster.NewCluster(a.Config, a.metadataService, a.executionChannel, &a.wg)
	a.cluster.Start()
	return nil
}

func (a *Agent) setupFlowEngine() error {
	a.flowEngine = engine.NewFlowEngine(a.cluster, a.executionChannel, a.metadataService, &a.wg)
	a.flowEngine.Start()
	return nil
}

func (a *Agent) setupWorkflowExecutionService() error {
	a.workflowExecutionService = service.NewWorkflowExecutionService(a.flowEngine)
	return nil
}

func (a *Agent) setupActionExecutorService() error {
	a.actionExecutionService = service.NewActionExecutionService(a.cluster, a.metadataService, a.flowEngine)
	return nil
}

func (a *Agent) setupHttpServer() error {
	var err error
	a.httpServer, err = rest.NewServer(a.Config.HttpPort, a.metadataService, a.workflowExecutionService)
	if err != nil {
		return err
	}
	return nil
}

func (a *Agent) setupGrpcServer() error {
	var err error
	conf := &rpc.GrpcConfig{
		ActionService:           a.actionExecutionService,
		ActionDefinitionService: a.metadataService.GetMetadataStorage(),
		GetServerer:             a.cluster.GetServerer(),
		ClusterRefresher:        a.cluster.GetClusterRefersher(),
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
		a.flowEngine.Stop,
		a.cluster.Stop,
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
