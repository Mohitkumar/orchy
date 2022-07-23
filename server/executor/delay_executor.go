package executor

import (
	"sync"

	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/persistence"
	"github.com/mohitkumar/orchy/server/util"
	"go.uber.org/zap"
)

var _ Executor = new(DelayExecutor)

type DelayExecutor struct {
	container     *container.DIContiner
	wg            *sync.WaitGroup
	stop          chan struct{}
	actionExector *ActionExecutor
}

func NewDelayExecutor(container *container.DIContiner, actionExector *ActionExecutor, wg *sync.WaitGroup) *DelayExecutor {
	return &DelayExecutor{
		container:     container,
		actionExector: actionExector,
		stop:          make(chan struct{}),
		wg:            wg,
	}
}

func (ex *DelayExecutor) Name() string {
	return "delay-executor"
}

func (ex *DelayExecutor) Start() error {
	fn := func() {
		res, err := ex.container.GetDelayQueue().Pop("delay_action")
		if err != nil {
			_, ok := err.(persistence.EmptyQueueError)
			if !ok {
				logger.Error("error while polling delay queue", zap.Error(err))
			}
			return
		}
		for _, r := range res {

			msg, err := ex.container.ActionExecutionRequestEncDec.Decode([]byte(r))
			if err != nil {
				logger.Error("can not decode action execution request")
				continue
			}
			err = ex.actionExector.Execute(*msg)
			if err != nil {
				logger.Error("error in executing workflow", zap.String("wfName", msg.WorkflowName), zap.String("flowId", msg.FlowId))
				continue
			}
		}
	}
	tw := util.NewTickWorker("delay-worker", 1, ex.stop, fn, ex.wg)
	tw.Start()
	logger.Info("delay executor started")
	return nil
}

func (ex *DelayExecutor) Stop() error {
	ex.stop <- struct{}{}
	return nil
}
