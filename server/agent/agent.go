package agent

import (
	"fmt"
	"net"
	"sync"

	"github.com/mohitkumar/orchy/server/config"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/executor"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/rest"
	"github.com/mohitkumar/orchy/server/rpc"
	"github.com/mohitkumar/orchy/server/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Agent struct {
	Config                   config.Config
	diContainer              *container.DIContiner
	httpServer               *rest.Server
	grpcServer               *grpc.Server
	delayExecutor            *executor.DelayExecutor
	actionExecutor           *executor.ActionExecutor
	retryExecutor            *executor.RetryExecutor
	timeoutExecutor          *executor.TimeoutExecutor
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
		a.setupDiContainer,
		a.setupActionExecutor,
		a.setupDelayExecutor,
		a.setupRetryExecutor,
		a.setupTimeoutExecutor,
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

func (a *Agent) setupDiContainer() error {
	a.diContainer = container.NewDiContainer()
	a.diContainer.Init(a.Config)
	return nil
}

func (a *Agent) setupActionExecutor() error {
	a.actionExecutor = executor.NewActionExecutor(a.diContainer, a.Config.ActionExecutorCapacity, &a.wg)
	return a.actionExecutor.Start()
}

func (a *Agent) setupDelayExecutor() error {
	a.delayExecutor = executor.NewDelayExecutor(a.diContainer, a.actionExecutor, &a.wg)
	return a.delayExecutor.Start()
}

func (a *Agent) setupRetryExecutor() error {
	a.retryExecutor = executor.NewRetryExecutor(a.diContainer, a.actionExecutor, &a.wg)
	return a.retryExecutor.Start()
}

func (a *Agent) setupTimeoutExecutor() error {
	a.timeoutExecutor = executor.NewTimeoutExecutor(a.diContainer, a.actionExecutor, &a.wg)
	return a.timeoutExecutor.Start()
}

func (a *Agent) setupWorkflowExecutionService() error {
	a.workflowExecutionService = service.NewWorkflowExecutionService(a.diContainer, a.actionExecutor)
	return nil
}

func (a *Agent) setupActionExecutorService() error {
	a.actionExecutionService = service.NewActionExecutionService(a.diContainer, a.actionExecutor)
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
		TaskService:    a.actionExecutionService,
		TaskDefService: a.diContainer.GetTaskDao(),
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
		a.actionExecutor.Stop,
		a.delayExecutor.Stop,
		a.retryExecutor.Stop,
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
