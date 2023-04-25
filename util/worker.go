package util

import (
	"sync"
	"time"

	"github.com/mohitkumar/orchy/logger"
	"go.uber.org/zap"
)

type Action any

type Worker struct {
	name    string
	stop    chan struct{}
	wg      *sync.WaitGroup
	handler func() bool
	running bool
}

func (w *Worker) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-w.stop:
				w.running = false
				logger.Info("stopping worker", zap.String("worker", w.name))
				return
			default:
				if !w.handler() {
					time.Sleep(1 * time.Second)
				}
			}
		}
	}()
	w.running = true
	logger.Info("executor started", zap.String("worker", w.name))
}

func (w *Worker) Stop() {
	w.stop <- struct{}{}
}

func (w *Worker) IsRunning() bool {
	return w.running
}

func NewWorker(name string, stop chan struct{}, handler func() bool, wg *sync.WaitGroup) *Worker {
	return &Worker{
		stop:    stop,
		wg:      wg,
		handler: handler,
		name:    name,
	}
}
