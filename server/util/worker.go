package util

import (
	"sync"

	"github.com/mohitkumar/orchy/server/logger"
	"go.uber.org/zap"
)

type Task any

type Worker struct {
	name     string
	capacity int
	stop     chan struct{}
	wg       *sync.WaitGroup
	handler  func(Task) error
	taskChan chan Task
}

func (w *Worker) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		for {
			select {
			case task := <-w.taskChan:
				err := w.handler(task)
				if err != nil {
					logger.Error("error in executing task in worker", zap.String("worker", w.name), zap.Any("task", task))
				}
			case <-w.stop:
				logger.Info("stopping worker", zap.String("worker", w.name))
				return
			}
		}
	}()
}

func (w *Worker) Sender() chan<- Task {
	return w.taskChan
}

func (w *Worker) Stop() {
	w.stop <- struct{}{}
}

func NewWorker(name string, wg *sync.WaitGroup, handler func(Task) error, capacity int) *Worker {
	ch := make(chan Task, capacity)
	stop := make(chan struct{})
	return &Worker{
		taskChan: ch,
		name:     name,
		wg:       wg,
		stop:     stop,
		handler:  handler,
	}
}
