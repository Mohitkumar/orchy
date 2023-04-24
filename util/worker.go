package util

import (
	"sync"

	"github.com/mohitkumar/orchy/logger"
	"go.uber.org/zap"
)

type Action any

type Worker struct {
	name       string
	capacity   int
	stop       chan struct{}
	wg         *sync.WaitGroup
	handler    func(Action) error
	actionChan chan Action
}

func (w *Worker) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		for {
			select {
			case action := <-w.actionChan:
				err := w.handler(action)
				if err != nil {
					logger.Error("error in executing action in worker", zap.String("worker", w.name), zap.Any("action", action), zap.Error(err))
				}
			case <-w.stop:
				logger.Info("stopping worker", zap.String("worker", w.name))
				return
			}
		}
	}()
}

func (w *Worker) Sender() chan<- Action {
	return w.actionChan
}

func (w *Worker) Stop() {
	w.stop <- struct{}{}
}

func NewWorker(name string, wg *sync.WaitGroup, handler func(Action) error, capacity int) *Worker {
	ch := make(chan Action, capacity)
	stop := make(chan struct{})
	return &Worker{
		actionChan: ch,
		name:       name,
		wg:         wg,
		stop:       stop,
		handler:    handler,
	}
}
