package util

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/server/logger"
	"go.uber.org/zap"
)

type TickWorker struct {
	stop         chan struct{}
	tickInterval time.Duration
	wg           *sync.WaitGroup
	name         string
	fn           func()
}

func NewTickWorker(name string, interval time.Duration, stop chan struct{}, fn func(), wg *sync.WaitGroup) *TickWorker {
	return &TickWorker{
		stop:         stop,
		tickInterval: interval,
		wg:           wg,
		fn:           fn,
		name:         name,
	}
}

func (tw *TickWorker) Start() {
	ticker := time.NewTicker(tw.tickInterval)
	tw.wg.Add(1)
	go func() {
		defer tw.wg.Done()
		for {
			select {
			case <-ticker.C:
				tw.fn()
			case <-tw.stop:
				logger.Info("stopping tick worker", zap.String("worker", tw.name))
				ticker.Stop()
				return
			}
		}
	}()
	logger.Info("executor started", zap.String("worker", tw.name))
}

func (tw *TickWorker) Stop() {
	tw.stop <- struct{}{}
}
