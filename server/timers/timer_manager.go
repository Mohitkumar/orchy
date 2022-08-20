package timers

import (
	"time"

	"github.com/RussellLuo/timingwheel"
)

type TimerManager struct {
	wheel *timingwheel.TimingWheel
}

func NewTimerManager(maxDelayInSeconds int64) *TimerManager {
	return &TimerManager{
		wheel: timingwheel.NewTimingWheel(time.Second, maxDelayInSeconds),
	}
}

func (m *TimerManager) AddTask(task func(), delaySeconds time.Duration) {
	m.wheel.AfterFunc(time.Second*delaySeconds, task)
}

func (m *TimerManager) Init() {
	m.wheel.Start()
}

func (m *TimerManager) Stop() {
	m.wheel.Stop()
}
