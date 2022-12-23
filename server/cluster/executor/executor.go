package executor

import (
	"sync"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
)

type Executor interface {
	Start()
	Stop()
	Name() string
}

type Executors struct {
	userExecutors    map[int]Executor
	systemExecutors  map[int]Executor
	delayExecutors   map[int]Executor
	retryExecutors   map[int]Executor
	timeoutExecutors map[int]Executor
	shards           *persistence.Shards
	partitions       int
	mu               sync.Mutex
}

func NewExecutors(shards *persistence.Shards) *Executors {
	return &Executors{
		userExecutors:    make(map[int]Executor),
		systemExecutors:  make(map[int]Executor),
		delayExecutors:   make(map[int]Executor),
		retryExecutors:   make(map[int]Executor),
		timeoutExecutors: make(map[int]Executor),
		shards:           shards,
	}
}

func (ex *Executors) InitExecutors(partitions int, container *container.DIContiner, wg *sync.WaitGroup) {
	ex.partitions = partitions
	for i := 0; i < partitions; i++ {
		ex.userExecutors[i] = NewUserActionExecutor(container, ex.shards.GetShard(i), wg)
		ex.systemExecutors[i] = NewSystemActionExecutor(container, ex.shards.GetShard(i), wg)
		ex.delayExecutors[i] = NewDelayExecutor(container, ex.shards.GetShard(i), wg)
		ex.retryExecutors[i] = NewRetryExecutor(container, ex.shards.GetShard(i), wg)
		ex.timeoutExecutors[i] = NewTimeoutExecutor(container, ex.shards.GetShard(i), wg)
	}
}

func (ex *Executors) Start(partition int) {
	ex.mu.Lock()
	defer ex.mu.Unlock()
	ex.userExecutors[partition].Start()
	ex.systemExecutors[partition].Start()
	ex.delayExecutors[partition].Start()
	ex.retryExecutors[partition].Start()
	ex.timeoutExecutors[partition].Start()
}

func (ex *Executors) Stop(partition int) {
	ex.mu.Lock()
	defer ex.mu.Unlock()
	ex.userExecutors[partition].Stop()
	ex.systemExecutors[partition].Stop()
	ex.delayExecutors[partition].Stop()
	ex.retryExecutors[partition].Stop()
	ex.timeoutExecutors[partition].Stop()
}

func (ex *Executors) StartAll() {
	ex.mu.Lock()
	defer ex.mu.Unlock()
	logger.Info("starting all executors")
	for i := 0; i < ex.partitions; i++ {
		ex.userExecutors[i].Start()
		ex.systemExecutors[i].Start()
		ex.delayExecutors[i].Start()
		ex.retryExecutors[i].Start()
		ex.timeoutExecutors[i].Start()
	}
}

func (ex *Executors) StopAll() error {
	ex.mu.Lock()
	defer ex.mu.Unlock()
	logger.Info("stoping all executors")
	for i := 0; i < ex.partitions; i++ {
		ex.userExecutors[i].Stop()
		ex.systemExecutors[i].Stop()
		ex.delayExecutors[i].Stop()
		ex.retryExecutors[i].Stop()
		ex.timeoutExecutors[i].Stop()
	}
	return nil
}
